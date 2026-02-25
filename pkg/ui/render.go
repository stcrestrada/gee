package ui

import (
	"fmt"
	"strings"
	"time"
)

type RepoResult struct {
	Name   string
	Stdout string
	Stderr string
	Failed bool
}

// RenderResults prints results for all repos with a summary footer.
// commandLabel is shown as a header (e.g., "$ git status", "clone").
// startTime is used to compute elapsed duration for the footer.
// Repos with no output and no error are collapsed to one line.
// Repos with output or errors get a bordered box.
func RenderResults(commandLabel string, results []RepoResult, startTime time.Time) {
	if commandLabel != "" {
		fmt.Println(StyleCommand.Render(commandLabel))
		fmt.Println()
	}

	successful := 0
	failed := 0

	for _, r := range results {
		hasStdout := strings.TrimSpace(r.Stdout) != ""
		hasStderr := strings.TrimSpace(r.Stderr) != ""

		if r.Failed {
			failed++
			renderExpanded(r, hasStdout, hasStderr)
		} else if hasStdout || hasStderr {
			successful++
			renderExpanded(r, hasStdout, hasStderr)
		} else {
			successful++
			fmt.Printf("%s %s\n", SymbolSuccess(), StyleRepoName.Render(r.Name))
		}
	}

	renderFooter(len(results), successful, failed, time.Since(startTime))
}

func renderExpanded(r RepoResult, hasStdout, hasStderr bool) {
	symbol := SymbolSuccess()
	boxStyle := StyleRepoBox
	if r.Failed {
		symbol = SymbolError()
		boxStyle = StyleRepoBoxError
	}

	fmt.Printf("%s %s\n", symbol, StyleRepoName.Render(r.Name))

	var sections []string
	if hasStdout {
		sections = append(sections, StyleStdout.Render(strings.TrimRight(r.Stdout, "\n")))
	}
	if hasStderr {
		label := StyleStderr.Render("stderr:")
		content := StyleStderr.Render(strings.TrimRight(r.Stderr, "\n"))
		sections = append(sections, label+"\n"+content)
	}

	if len(sections) > 0 {
		body := strings.Join(sections, "\n")
		fmt.Println(boxStyle.Render(body))
	}
	fmt.Println()
}

func renderFooter(total, successful, failed int, elapsed time.Duration) {
	seconds := elapsed.Seconds()

	parts := []string{
		fmt.Sprintf("Total: %d", total),
		StyleSuccess.Render(fmt.Sprintf("Successful: %d", successful)),
		StyleError.Render(fmt.Sprintf("Failed: %d", failed)),
		StyleSummaryLine.Render(fmt.Sprintf("Time: %.1fs", seconds)),
	}

	line := strings.Join(parts, StyleSummaryLine.Render(" | "))
	fmt.Println(StyleFooter.Render(line))
}
