package tui

import (
	"gee/pkg/command"
	"gee/pkg/types"
	"gee/pkg/ui"
	"gee/pkg/util"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// View identifies which screen is active.
type View int

const (
	ViewDashboard View = iota
	ViewDiscovery
)

// RepoRow holds display state for one repo in the dashboard table.
type RepoRow struct {
	Repo    types.Repo
	Status  ui.StatusSummary
	Failed  bool
	Loading bool
	Action  string // "pulling...", "exec...", or ""
}

// DiscoveryModel holds state for the remote discovery view.
type DiscoveryModel struct {
	Provider    string
	RemoteRepos []RemoteRepo
	Selected    map[int]bool
	Cursor      int
	Loading     bool
	Error       error
}

// AppModel is the root bubbletea model.
type AppModel struct {
	// Core
	Config    *types.GeeContext
	RepoUtils *util.RepoUtils
	Git       command.GitRepoOperation
	Caps      Capabilities

	// View routing
	ActiveView View

	// Dashboard
	Rows      []RepoRow
	Cursor    int
	StatusCh  <-chan StatusResultMsg // held between Update calls to drain pool results
	Filter    string
	Filtering bool
	FilterInput textinput.Model

	// Exec overlay
	ExecInput  textinput.Model
	ExecActive bool

	// Action log (recent results shown at bottom)
	ActionLog []string

	// Discovery
	Discovery DiscoveryModel

	// Terminal dimensions
	Width  int
	Height int

	// State
	Refreshing bool
}

// NewAppModel creates a ready-to-use AppModel from a loaded config.
func NewAppModel(config *types.GeeContext) AppModel {
	git := command.GitRepoOperation{}
	repoUtils := util.NewRepoUtils(git)

	rows := make([]RepoRow, len(config.Repos))
	for i, repo := range config.Repos {
		rows[i] = RepoRow{Repo: repo, Loading: true}
	}

	caps := DetectCapabilities()

	filterInput := textinput.New()
	filterInput.Placeholder = "filter repos..."
	filterInput.CharLimit = 64

	execInput := textinput.New()
	execInput.Placeholder = "command to run..."
	execInput.CharLimit = 256

	return AppModel{
		Config:      config,
		RepoUtils:   repoUtils,
		Git:         git,
		Caps:        caps,
		Rows:        rows,
		FilterInput: filterInput,
		ExecInput:   execInput,
		Discovery: DiscoveryModel{
			Provider: DiscoveryProvider(caps),
			Selected: make(map[int]bool),
		},
	}
}

// Init kicks off the initial status refresh and periodic tick.
func (m AppModel) Init() tea.Cmd {
	cmd, ch := refreshStatusCmd(m.Config.Repos, m.RepoUtils)
	m.StatusCh = ch
	m.Refreshing = true
	// We cannot mutate m here (Init returns tea.Cmd, not tea.Model).
	// So we use a wrapper cmd that sets the channel on first StatusResultMsg.
	// Actually, bubbletea Init just returns a Cmd — we store ch via a
	// bootstrap message instead.
	return tea.Batch(cmd, tickCmd(), func() tea.Msg {
		return initStatusChanMsg{ch: ch}
	})
}

// initStatusChanMsg is an internal message to hand the status channel to Update
// since Init cannot mutate the model directly.
type initStatusChanMsg struct {
	ch <-chan StatusResultMsg
}
