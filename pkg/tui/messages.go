package tui

import "gee/pkg/ui"

// StatusResultMsg delivers one repo's porcelain status result into the
// bubbletea Update loop. Sent once per repo during a status refresh.
type StatusResultMsg struct {
	Index   int
	Name    string
	Summary ui.StatusSummary
	Failed  bool
}

// StatusRefreshDoneMsg signals that every repo has reported status and
// the refresh channel is closed.
type StatusRefreshDoneMsg struct{}

// PullResultMsg delivers the result of a pull on a single repo.
type PullResultMsg struct {
	Index  int
	Name   string
	Stdout string
	Stderr string
	Failed bool
}

// ExecResultMsg delivers the result of an exec on a single repo.
type ExecResultMsg struct {
	Index  int
	Name   string
	Stdout string
	Stderr string
	Failed bool
}

// RemoteRepo represents a repository discovered from gh or glab.
type RemoteRepo struct {
	FullName    string
	Description string
	CloneURL    string
	Private     bool
}

// DiscoveryResultMsg delivers the list of remote repos from gh/glab.
type DiscoveryResultMsg struct {
	Repos []RemoteRepo
	Error error
}

// CloneBatchResultMsg reports one cloned repo's outcome.
type CloneBatchResultMsg struct {
	Name   string
	Failed bool
}

// CloneBatchDoneMsg signals batch clone is complete.
type CloneBatchDoneMsg struct {
	Succeeded int
	Failed    int
}

// TickMsg triggers periodic status refresh.
type TickMsg struct{}
