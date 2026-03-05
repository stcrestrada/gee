package tui

import (
	"path/filepath"

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
	Pinned  bool
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
	// Core — cache is the single source of truth
	Cache     *util.RepoCache
	RepoUtils *util.RepoUtils
	Git       command.GitRepoOperation
	Caps      Capabilities

	// View routing
	ActiveView View

	// Dashboard
	Rows        []RepoRow
	Cursor      int
	StatusCh    <-chan StatusResultMsg
	ScanCh      <-chan RepoDiscoveredMsg
	Filter      string
	Filtering   bool
	FilterInput textinput.Model

	// Exec overlay
	ExecInput  textinput.Model
	ExecActive bool

	// Action log (recent results shown at bottom)
	ActionLog []string

	// Discovery (remote — gh/glab)
	Discovery DiscoveryModel

	// Terminal dimensions
	Width  int
	Height int

	// State
	Refreshing bool
	Scanning   bool

	// Teleport — set when user presses Enter, read by main after Run()
	SelectedPath string
}

// NewAppModel creates a ready-to-use AppModel from the cache.
func NewAppModel(cache *util.RepoCache) AppModel {
	git := command.GitRepoOperation{}
	repoUtils := util.NewRepoUtils(git)

	// Seed rows from cache for instant startup.
	cached := cache.All()
	rows := make([]RepoRow, len(cached))
	for i, c := range cached {
		rows[i] = RepoRow{
			Repo: types.Repo{
				Name:   c.Name,
				Path:   filepath.Dir(c.Path),
				Remote: c.Remote,
			},
			Pinned:  c.Pinned,
			Loading: true,
		}
	}

	caps := DetectCapabilities()

	filterInput := textinput.New()
	filterInput.Placeholder = "filter repos..."
	filterInput.CharLimit = 64

	execInput := textinput.New()
	execInput.Placeholder = "command to run..."
	execInput.CharLimit = 256

	return AppModel{
		Cache:       cache,
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

// Init kicks off the initial status refresh, background scanner, and periodic tick.
func (m AppModel) Init() tea.Cmd {
	repos := m.repoSlice()
	statusCmd, statusCh := refreshStatusCmd(repos, m.RepoUtils)
	scanCmd, scanCh := scanLocalReposCmd(m.Cache)

	return tea.Batch(
		statusCmd,
		scanCmd,
		tickCmd(),
		func() tea.Msg { return initStatusChanMsg{ch: statusCh} },
		func() tea.Msg { return initScanChanMsg{ch: scanCh} },
	)
}

// repoSlice extracts []types.Repo from the current row state.
func (m *AppModel) repoSlice() []types.Repo {
	repos := make([]types.Repo, len(m.Rows))
	for i, r := range m.Rows {
		repos[i] = r.Repo
	}
	return repos
}

// initStatusChanMsg is an internal message to hand the status channel to Update
// since Init cannot mutate the model directly.
type initStatusChanMsg struct {
	ch <-chan StatusResultMsg
}
