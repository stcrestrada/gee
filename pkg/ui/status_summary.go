package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StatusSummary struct {
	Branch    string
	State     string // "", "REBASE", "MERGE", "CHERRY-PICK"
	Progress  string // e.g. "3/5" for rebase, empty otherwise
	Ahead     int
	Behind    int
	Staged    int
	Modified  int
	Untracked int
	Conflicts int
	Stale     bool // true if dirty with newest top-level file mtime > 7 days
}

// ParsePorcelainV2 parses `git status --porcelain=v2 --branch` output.
func ParsePorcelainV2(output string) StatusSummary {
	s := StatusSummary{}
	for _, line := range strings.Split(output, "\n") {
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			s.Branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.ab "):
			fmt.Sscanf(strings.TrimPrefix(line, "# branch.ab "), "+%d -%d", &s.Ahead, &s.Behind)
		case strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 "):
			// Changed entry: XY field is at index 1
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				xy := fields[1]
				if len(xy) == 2 {
					if xy[0] != '.' {
						s.Staged++
					}
					if xy[1] != '.' {
						s.Modified++
					}
				}
			}
		case strings.HasPrefix(line, "u "):
			// Unmerged entry (conflict)
			s.Conflicts++
		case strings.HasPrefix(line, "? "):
			s.Untracked++
		}
	}
	return s
}

// DetectGitState checks .git/ sentinel files to identify rebase, merge,
// or cherry-pick in progress. repoPath is the repo's working directory.
func DetectGitState(repoPath string) (state string, progress string) {
	gitDir := filepath.Join(repoPath, ".git")

	// Interactive rebase: .git/rebase-merge/
	if info, err := os.Stat(filepath.Join(gitDir, "rebase-merge")); err == nil && info.IsDir() {
		state = "REBASE"
		msgnum, err1 := os.ReadFile(filepath.Join(gitDir, "rebase-merge", "msgnum"))
		end, err2 := os.ReadFile(filepath.Join(gitDir, "rebase-merge", "end"))
		if err1 == nil && err2 == nil {
			progress = fmt.Sprintf("%s/%s",
				strings.TrimSpace(string(msgnum)),
				strings.TrimSpace(string(end)))
		}
		return
	}

	// git am or non-interactive rebase: .git/rebase-apply/
	if info, err := os.Stat(filepath.Join(gitDir, "rebase-apply")); err == nil && info.IsDir() {
		state = "REBASE"
		next, err1 := os.ReadFile(filepath.Join(gitDir, "rebase-apply", "next"))
		last, err2 := os.ReadFile(filepath.Join(gitDir, "rebase-apply", "last"))
		if err1 == nil && err2 == nil {
			progress = fmt.Sprintf("%s/%s",
				strings.TrimSpace(string(next)),
				strings.TrimSpace(string(last)))
		}
		return
	}

	// Merge in progress
	if _, err := os.Stat(filepath.Join(gitDir, "MERGE_HEAD")); err == nil {
		state = "MERGE"
		return
	}

	// Cherry-pick in progress
	if _, err := os.Stat(filepath.Join(gitDir, "CHERRY_PICK_HEAD")); err == nil {
		state = "CHERRY-PICK"
		return
	}

	return "", ""
}

// CheckStaleness returns true if the repo has uncommitted changes and the
// most recent top-level file's mtime is older than 7 days. This is a fast
// heuristic — only top-level entries are checked, not a deep walk.
func CheckStaleness(repoPath string, s StatusSummary) bool {
	if s.Modified+s.Staged+s.Untracked == 0 {
		return false
	}
	var newest time.Time
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
	}
	if newest.IsZero() {
		return false
	}
	return time.Since(newest) > 7*24*time.Hour
}

type RepoStatusResult struct {
	Name    string
	Summary StatusSummary
	Failed  bool
}

// RenderStatusTable prints a compact one-line-per-repo status dashboard
// with a telemetry footer.
func RenderStatusTable(results []RepoStatusResult, startTime time.Time) {
	successful := 0
	failed := 0

	for _, r := range results {
		if r.Failed {
			failed++
			fmt.Printf("%s %s  %s\n", SymbolError(), StyleRepoName.Render(r.Name), StyleError.Render("failed"))
			continue
		}

		successful++
		s := r.Summary
		parts := []string{StyleRepoName.Render(r.Name)}

		// Branch with state annotation
		branchDisplay := s.Branch
		if s.State != "" {
			if s.Progress != "" {
				branchDisplay = fmt.Sprintf("%s|%s %s", s.Branch, s.State, s.Progress)
			} else {
				branchDisplay = fmt.Sprintf("%s|%s", s.Branch, s.State)
			}
			parts = append(parts, StyleWarning.Render(branchDisplay))
		} else {
			parts = append(parts, StyleCommand.Render(branchDisplay))
		}

		// Ahead/behind
		if s.Ahead > 0 || s.Behind > 0 {
			ab := ""
			if s.Ahead > 0 {
				ab += StyleSuccess.Render(fmt.Sprintf("↑%d", s.Ahead))
			}
			if s.Behind > 0 {
				if ab != "" {
					ab += " "
				}
				ab += StyleError.Render(fmt.Sprintf("↓%d", s.Behind))
			}
			parts = append(parts, ab)
		}

		// File changes
		changes := []string{}
		if s.Conflicts > 0 {
			changes = append(changes, StyleError.Render(fmt.Sprintf("!%d conflict", s.Conflicts)))
		}
		if s.Staged > 0 {
			changes = append(changes, StyleSuccess.Render(fmt.Sprintf("+%d staged", s.Staged)))
		}
		if s.Modified > 0 {
			changes = append(changes, StyleWarning.Render(fmt.Sprintf("~%d modified", s.Modified)))
		}
		if s.Untracked > 0 {
			changes = append(changes, StyleCommand.Render(fmt.Sprintf("?%d untracked", s.Untracked)))
		}

		if len(changes) == 0 {
			parts = append(parts, StyleSuccess.Render("clean"))
		} else {
			parts = append(parts, strings.Join(changes, " "))
		}

		fmt.Printf("%s  %s\n", SymbolSuccess(), strings.Join(parts, "  "))
	}

	renderFooter(len(results), successful, failed, time.Since(startTime))
}
