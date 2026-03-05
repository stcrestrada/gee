package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gee/pkg/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// filteredRow pairs a RepoRow with its original index in m.Rows.
type filteredRow struct {
	origIndex int
	row       RepoRow
}

// filteredRows returns the subset of Rows matching the current filter.
func (m *AppModel) filteredRows() []filteredRow {
	if m.Filter == "" {
		rows := make([]filteredRow, len(m.Rows))
		for i, r := range m.Rows {
			rows[i] = filteredRow{origIndex: i, row: r}
		}
		return rows
	}
	lowerFilter := strings.ToLower(m.Filter)
	var rows []filteredRow
	for i, r := range m.Rows {
		if strings.Contains(strings.ToLower(r.Repo.Name), lowerFilter) {
			rows = append(rows, filteredRow{origIndex: i, row: r})
		}
	}
	return rows
}

// startRefresh begins a new status refresh cycle, returning the tea.Cmd to
// start draining results.
func (m *AppModel) startRefresh() tea.Cmd {
	m.Refreshing = true
	for i := range m.Rows {
		m.Rows[i].Loading = true
	}
	repos := m.repoSlice()
	cmd, ch := refreshStatusCmd(repos, m.RepoUtils)
	m.StatusCh = ch
	return cmd
}

// reloadCache re-reads cache from disk and rebuilds the row list.
func (m *AppModel) reloadCache() {
	if _, err := m.Cache.Load(); err != nil {
		m.ActionLog = append(m.ActionLog, fmt.Sprintf("reload cache: %s", err))
		return
	}
	cached := m.Cache.All()

	// Preserve status for repos that still exist.
	oldByPath := make(map[string]RepoRow, len(m.Rows))
	for _, r := range m.Rows {
		fullPath := m.RepoUtils.FullPathWithRepo(r.Repo.Path, r.Repo.Name)
		oldByPath[fullPath] = r
	}
	rows := make([]RepoRow, len(cached))
	for i, c := range cached {
		if old, ok := oldByPath[c.Path]; ok {
			old.Pinned = c.Pinned
			rows[i] = old
		} else {
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
	}
	m.Rows = rows
}

// Update is the main bubbletea update function.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	// Bootstrap: store the status channel from Init.
	case initStatusChanMsg:
		m.StatusCh = msg.ch
		m.Refreshing = true
		return m, nil

	// Bootstrap: store the scanner channel from Init.
	case initScanChanMsg:
		m.ScanCh = msg.ch
		m.Scanning = true
		return m, nil

	// --- Status refresh stream ---
	case StatusResultMsg:
		if msg.Index >= 0 && msg.Index < len(m.Rows) {
			m.Rows[msg.Index].Loading = false
			m.Rows[msg.Index].Failed = msg.Failed
			if !msg.Failed {
				m.Rows[msg.Index].Status = msg.Summary
			}
		}
		// Keep draining the channel.
		if m.StatusCh != nil {
			return m, waitForStatusResult(m.StatusCh)
		}
		return m, nil

	case StatusRefreshDoneMsg:
		m.Refreshing = false
		m.StatusCh = nil
		return m, nil

	// --- Scanner stream ---
	case RepoDiscoveredMsg:
		newRow := RepoRow{
			Repo: types.Repo{
				Name:   msg.Name,
				Path:   filepath.Dir(msg.Path),
				Remote: msg.Remote,
			},
			Pinned:  false,
			Loading: true,
		}
		m.Rows = append(m.Rows, newRow)
		newIndex := len(m.Rows) - 1

		statusCmd := refreshSingleRepoStatusCmd(m.Rows[newIndex].Repo, newIndex, m.RepoUtils)
		var nextScanCmd tea.Cmd
		if m.ScanCh != nil {
			nextScanCmd = waitForDiscoveredRepo(m.ScanCh)
		}
		return m, tea.Batch(statusCmd, nextScanCmd)

	case ScanDoneMsg:
		m.Scanning = false
		m.ScanCh = nil
		m.ActionLog = append(m.ActionLog, fmt.Sprintf("scan complete: %d repos", len(m.Rows)))
		return m, nil

	// --- Periodic refresh ---
	case TickMsg:
		if !m.Refreshing {
			return m, tea.Batch(m.startRefresh(), tickCmd())
		}
		return m, tickCmd()

	// --- Pull result ---
	case PullResultMsg:
		if msg.Index >= 0 && msg.Index < len(m.Rows) {
			m.Rows[msg.Index].Action = ""
		}
		if msg.Failed {
			stderr := strings.TrimSpace(msg.Stderr)
			if stderr == "" {
				stderr = "unknown error"
			}
			m.ActionLog = append(m.ActionLog, fmt.Sprintf("pull %s: FAILED - %s", msg.Name, stderr))
		} else {
			out := strings.TrimSpace(msg.Stdout)
			if out == "" {
				out = "up to date"
			}
			m.ActionLog = append(m.ActionLog, fmt.Sprintf("pull %s: %s", msg.Name, out))
		}
		return m, m.startRefresh()

	// --- Exec result ---
	case ExecResultMsg:
		if msg.Index >= 0 && msg.Index < len(m.Rows) {
			m.Rows[msg.Index].Action = ""
		}
		out := strings.TrimSpace(msg.Stdout)
		if msg.Failed {
			out = strings.TrimSpace(msg.Stderr)
			if out == "" {
				out = "failed"
			}
			m.ActionLog = append(m.ActionLog, fmt.Sprintf("exec %s: FAILED - %s", msg.Name, out))
		} else {
			if out == "" {
				out = "done"
			}
			m.ActionLog = append(m.ActionLog, fmt.Sprintf("exec %s: %s", msg.Name, truncate(out, 80)))
		}
		return m, m.startRefresh()

	// --- Discovery ---
	case DiscoveryResultMsg:
		m.Discovery.Loading = false
		m.Discovery.Error = msg.Error
		m.Discovery.RemoteRepos = msg.Repos
		return m, nil

	case CloneBatchDoneMsg:
		m.ActionLog = append(m.ActionLog,
			fmt.Sprintf("clone: %d succeeded, %d failed", msg.Succeeded, msg.Failed))
		m.reloadCache()
		m.ActiveView = ViewDashboard
		return m, m.startRefresh()

	// --- Keyboard input ---
	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		switch m.ActiveView {
		case ViewDashboard:
			return m.updateDashboard(msg)
		case ViewDiscovery:
			return m.updateDiscovery(msg)
		}
	}

	return m, nil
}

func (m AppModel) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// --- Exec input mode ---
	if m.ExecActive {
		switch msg.String() {
		case "enter":
			userCmd := m.ExecInput.Value()
			m.ExecInput.Reset()
			m.ExecActive = false
			if userCmd == "" {
				return m, nil
			}
			filtered := m.filteredRows()
			if m.Cursor >= 0 && m.Cursor < len(filtered) {
				r := filtered[m.Cursor]
				m.Rows[r.origIndex].Action = "exec..."
				return m, execRepoCmd(r.row.Repo, r.origIndex, userCmd, m.RepoUtils)
			}
			return m, nil
		case "esc":
			m.ExecInput.Reset()
			m.ExecActive = false
			return m, nil
		default:
			var cmd tea.Cmd
			m.ExecInput, cmd = m.ExecInput.Update(msg)
			return m, cmd
		}
	}

	// --- Filter input mode ---
	if m.Filtering {
		switch msg.String() {
		case "enter":
			m.Filtering = false
			m.FilterInput.Blur()
			m.Filter = m.FilterInput.Value()
			m.Cursor = 0
			return m, nil
		case "esc":
			m.Filtering = false
			m.FilterInput.Blur()
			m.FilterInput.SetValue("")
			m.Filter = ""
			m.Cursor = 0
			return m, nil
		default:
			var cmd tea.Cmd
			m.FilterInput, cmd = m.FilterInput.Update(msg)
			m.Filter = m.FilterInput.Value()
			return m, cmd
		}
	}

	// --- Normal dashboard navigation ---
	filtered := m.filteredRows()
	maxIdx := len(filtered) - 1
	if maxIdx < 0 {
		maxIdx = 0
	}

	switch msg.String() {
	case "j", "down":
		if m.Cursor < maxIdx {
			m.Cursor++
		}
	case "k", "up":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "g":
		m.Cursor = 0
	case "G":
		m.Cursor = maxIdx

	case "p":
		if len(filtered) > 0 && m.Cursor <= maxIdx {
			r := filtered[m.Cursor]
			m.Rows[r.origIndex].Action = "pulling..."
			return m, pullRepoCmd(r.row.Repo, r.origIndex, m.RepoUtils)
		}

	case "P":
		var cmds []tea.Cmd
		for _, r := range filtered {
			m.Rows[r.origIndex].Action = "pulling..."
			cmds = append(cmds, pullRepoCmd(r.row.Repo, r.origIndex, m.RepoUtils))
		}
		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}

	case "e":
		m.ExecActive = true
		m.ExecInput.Focus()
		return m, textinput.Blink

	case "enter":
		if len(filtered) > 0 && m.Cursor <= maxIdx {
			r := filtered[m.Cursor]
			fullPath := m.RepoUtils.FullPathWithRepo(r.row.Repo.Path, r.row.Repo.Name)
			return m, openShellCmd(fullPath)
		}

	case "a":
		// Toggle pin on selected repo.
		if len(filtered) > 0 && m.Cursor <= maxIdx {
			r := filtered[m.Cursor]
			fullPath := m.RepoUtils.FullPathWithRepo(r.row.Repo.Path, r.row.Repo.Name)
			if m.Rows[r.origIndex].Pinned {
				m.Cache.Unpin(fullPath)
				m.Rows[r.origIndex].Pinned = false
				m.ActionLog = append(m.ActionLog, fmt.Sprintf("unpinned %s", r.row.Repo.Name))
			} else {
				m.Cache.Pin(fullPath)
				m.Rows[r.origIndex].Pinned = true
				m.ActionLog = append(m.ActionLog, fmt.Sprintf("pinned %s", r.row.Repo.Name))
			}
			m.Cache.Save()
		}

	case "r":
		if !m.Refreshing {
			return m, m.startRefresh()
		}

	case "/":
		m.Filtering = true
		m.FilterInput.Focus()
		return m, textinput.Blink

	case "d":
		if m.Discovery.Provider != "" {
			m.ActiveView = ViewDiscovery
			if len(m.Discovery.RemoteRepos) == 0 && !m.Discovery.Loading {
				m.Discovery.Loading = true
				return m, discoverRemoteReposCmd(m.Discovery.Provider)
			}
		}

	case "q":
		return m, tea.Quit
	}

	return m, nil
}

func (m AppModel) updateDiscovery(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIdx := len(m.Discovery.RemoteRepos) - 1
	if maxIdx < 0 {
		maxIdx = 0
	}

	switch msg.String() {
	case "esc", "d":
		m.ActiveView = ViewDashboard
	case "j", "down":
		if m.Discovery.Cursor < maxIdx {
			m.Discovery.Cursor++
		}
	case "k", "up":
		if m.Discovery.Cursor > 0 {
			m.Discovery.Cursor--
		}
	case "g":
		m.Discovery.Cursor = 0
	case "G":
		m.Discovery.Cursor = maxIdx
	case " ":
		m.Discovery.Selected[m.Discovery.Cursor] = !m.Discovery.Selected[m.Discovery.Cursor]
		if m.Discovery.Cursor < maxIdx {
			m.Discovery.Cursor++
		}
	case "enter":
		var toClone []RemoteRepo
		for i, sel := range m.Discovery.Selected {
			if sel && i < len(m.Discovery.RemoteRepos) {
				toClone = append(toClone, m.Discovery.RemoteRepos[i])
			}
		}
		if len(toClone) == 0 {
			return m, nil
		}
		m.Discovery.Loading = true
		cloneDir, _ := os.UserHomeDir()
		return m, cloneBatchCmd(toClone, m.Cache, cloneDir, m.RepoUtils)
	case "q":
		return m, tea.Quit
	}

	return m, nil
}

func truncate(s string, maxLen int) string {
	if idx := strings.IndexByte(s, '\n'); idx != -1 {
		s = s[:idx]
	}
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
