package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gee/pkg/command"
	"gee/pkg/types"
	"gee/pkg/ui"
	"gee/pkg/util"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stcrestrada/gogo/v3"
)

// refreshStatusCmd spawns a gogo pool that runs `git status --porcelain=v2 --branch`
// on every repo concurrently. Results are fed into a Go channel; the returned
// tea.Cmd reads one result at a time, yielding StatusResultMsg back to Update.
// When the channel closes, StatusRefreshDoneMsg is returned.
//
// This is the core gogo-to-bubbletea bridge pattern:
//   1. gogo pool writes to chan StatusResultMsg
//   2. tea.Cmd blocks reading one value from that channel
//   3. Update processes the msg and returns another tea.Cmd to drain the next value
//   4. When channel closes → StatusRefreshDoneMsg
func refreshStatusCmd(repos []types.Repo, repoUtils *util.RepoUtils) (tea.Cmd, <-chan StatusResultMsg) {
	git := command.GitRepoOperation{}
	ch := make(chan StatusResultMsg, len(repos))

	pool := gogo.NewPool[struct{}](
		context.Background(),
		len(repos),
		len(repos),
		func(ctx context.Context, i int) (struct{}, error) {
			repo := repos[i]
			fullPath := repoUtils.FullPathWithRepo(repo.Path, repo.Name)
			rc := &types.RunConfig{
				StdOut: &bytes.Buffer{},
				StdErr: &bytes.Buffer{},
			}

			msg := StatusResultMsg{Index: i, Name: repo.Name}

			git.StatusPorcelain(repo.Name, fullPath, rc, func(onFinish *types.CommandOnFinish) {
				if onFinish.Failed {
					msg.Failed = true
				} else {
					summary := ui.ParsePorcelainV2(rc.StdOut.String())
					if summary.Branch == "(detached)" {
						summary.State, summary.Progress = ui.DetectGitState(fullPath)
					}
					msg.Summary = summary
				}
			})

			ch <- msg
			return struct{}{}, nil
		},
	)

	// Drain pool results in background, then close the channel.
	go func() {
		for range pool.Go() {
		}
		close(ch)
	}()

	return waitForStatusResult(ch), ch
}

// waitForStatusResult returns a tea.Cmd that reads one StatusResultMsg from
// the channel. When the channel is closed it returns StatusRefreshDoneMsg.
// After each successful read, Update should call this again with the same
// channel to keep draining.
func waitForStatusResult(ch <-chan StatusResultMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return StatusRefreshDoneMsg{}
		}
		return msg
	}
}

// pullRepoCmd pulls a single repo and returns the result.
func pullRepoCmd(repo types.Repo, index int, repoUtils *util.RepoUtils) tea.Cmd {
	return func() tea.Msg {
		git := command.GitRepoOperation{}
		fullPath := repoUtils.FullPathWithRepo(repo.Path, repo.Name)
		rc := &types.RunConfig{
			StdOut: &bytes.Buffer{},
			StdErr: &bytes.Buffer{},
		}

		msg := PullResultMsg{Index: index, Name: repo.Name}
		git.Pull(repo.Name, fullPath, rc, func(onFinish *types.CommandOnFinish) {
			msg.Stdout = rc.StdOut.String()
			msg.Stderr = rc.StdErr.String()
			msg.Failed = onFinish.Failed
		})
		return msg
	}
}

// execRepoCmd runs an arbitrary shell command in a single repo directory.
func execRepoCmd(repo types.Repo, index int, userCmd string, repoUtils *util.RepoUtils) tea.Cmd {
	return func() tea.Msg {
		fullPath := repoUtils.FullPathWithRepo(repo.Path, repo.Name)
		var stdout, stderr bytes.Buffer
		sh := exec.Command("sh", "-c", userCmd)
		sh.Dir = fullPath
		sh.Stdout = &stdout
		sh.Stderr = &stderr
		err := sh.Run()

		return ExecResultMsg{
			Index:  index,
			Name:   repo.Name,
			Stdout: stdout.String(),
			Stderr: stderr.String(),
			Failed: err != nil,
		}
	}
}

// tickCmd returns a tea.Cmd that fires a TickMsg after the refresh interval.
func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// openShellCmd suspends the TUI and opens the user's shell in the given directory.
// On exit, it triggers a TickMsg to refresh status.
func openShellCmd(dir string) tea.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	c := exec.Command(shell)
	c.Dir = dir
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return TickMsg{}
	})
}

// discoverRemoteReposCmd calls gh or glab to list remote repos.
func discoverRemoteReposCmd(provider string) tea.Cmd {
	return func() tea.Msg {
		switch provider {
		case "gh":
			return discoverGitHub()
		case "glab":
			return discoverGitLab()
		default:
			return DiscoveryResultMsg{Error: fmt.Errorf("no discovery provider available")}
		}
	}
}

func discoverGitHub() tea.Msg {
	cmd := exec.Command("gh", "repo", "list",
		"--json", "nameWithOwner,description,sshUrl,isPrivate",
		"--limit", "100")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		return DiscoveryResultMsg{Error: fmt.Errorf("gh repo list: %w", err)}
	}

	var raw []struct {
		NameWithOwner string `json:"nameWithOwner"`
		Description   string `json:"description"`
		SSHURL        string `json:"sshUrl"`
		IsPrivate     bool   `json:"isPrivate"`
	}
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
		return DiscoveryResultMsg{Error: fmt.Errorf("parse gh output: %w", err)}
	}

	repos := make([]RemoteRepo, len(raw))
	for i, r := range raw {
		repos[i] = RemoteRepo{
			FullName:    r.NameWithOwner,
			Description: r.Description,
			CloneURL:    r.SSHURL,
			Private:     r.IsPrivate,
		}
	}
	return DiscoveryResultMsg{Repos: repos}
}

func discoverGitLab() tea.Msg {
	cmd := exec.Command("glab", "repo", "list", "-O", "json")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		return DiscoveryResultMsg{Error: fmt.Errorf("glab repo list: %w", err)}
	}

	var raw []struct {
		PathWithNamespace string `json:"path_with_namespace"`
		Description       string `json:"description"`
		SSHURLToRepo      string `json:"ssh_url_to_repo"`
		Visibility        string `json:"visibility"`
	}
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
		return DiscoveryResultMsg{Error: fmt.Errorf("parse glab output: %w", err)}
	}

	repos := make([]RemoteRepo, len(raw))
	for i, r := range raw {
		repos[i] = RemoteRepo{
			FullName:    r.PathWithNamespace,
			Description: r.Description,
			CloneURL:    r.SSHURLToRepo,
			Private:     r.Visibility == "private",
		}
	}
	return DiscoveryResultMsg{Repos: repos}
}

// cloneBatchCmd clones a set of remote repos using gogo, adds them to the
// config, and returns a CloneBatchDoneMsg.
func cloneBatchCmd(toClone []RemoteRepo, config *types.GeeContext, repoUtils *util.RepoUtils) tea.Cmd {
	return func() tea.Msg {
		git := command.GitRepoOperation{}
		configDir := config.ConfigFilePath
		succeeded := 0
		failed := 0

		pool := gogo.NewPool[struct{}](
			context.Background(),
			len(toClone),
			len(toClone),
			func(ctx context.Context, i int) (struct{}, error) {
				remote := toClone[i]
				rc := &types.RunConfig{
					StdOut: &bytes.Buffer{},
					StdErr: &bytes.Buffer{},
				}

				// Extract repo name from "owner/name" → "name"
				parts := strings.Split(remote.FullName, "/")
				repoName := parts[len(parts)-1]

				var cloneFailed bool
				git.Clone(repoName, remote.CloneURL, configDir, rc, func(onFinish *types.CommandOnFinish) {
					cloneFailed = onFinish.Failed
					// "already exists" is not a real failure
					if cloneFailed && strings.Contains(rc.StdErr.String(), "already exists") {
						cloneFailed = false
					}
				})

				if cloneFailed {
					failed++
				} else {
					succeeded++
					// Add to config
					config.Repos = append(config.Repos, types.Repo{
						Name:   repoName,
						Path:   configDir,
						Remote: remote.CloneURL,
					})
				}
				return struct{}{}, nil
			},
		)

		pool.Wait()

		// Save updated config
		if succeeded > 0 {
			util.NewConfigHelper().SaveConfig(config)
		}

		return CloneBatchDoneMsg{Succeeded: succeeded, Failed: failed}
	}
}
