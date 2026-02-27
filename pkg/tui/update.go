package tui

import (
	"fmt"
	"strings"

	"gee/pkg/util"

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
	cmd, ch := refreshStatusCmd(m.Config.Repos, m.RepoUtils)
	m.StatusCh = ch
	return cmd
}

// reloadConfig re-reads gee.toml from disk and rebuilds the row list.
func (m *AppModel) reloadConfig() {
	helper := util.NewConfigHelper()
	config, err := helper.LoadConfig(m.Config.ConfigFilePath)
	if err != nil {
		m.ActionLog = append(m.ActionLog, fmt.Sprintf("reload config: %s", err))
		return
	}
	m.Config = config

	// Rebuild rows, preserving status for repos that still exist.
	oldByName := make(map[string]RepoRow, len(m.Rows))
	for _, r := range m.Rows {
		oldByName[r.Repo.Name] = r
	}
	rows := make([]RepoRow, len(config.Repos))
	for i, repo := range config.Repos {
		if old, ok := oldByName[repo.Name]; ok {
			old.Repo = repo
			rows[i] = old
		} else {
			rows[i] = RepoRow{Repo: repo, Loading: true}
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
		m.reloadConfig()
		// Switch back to dashboard and refresh.
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
		// Pull selected repo.
		if len(filtered) > 0 && m.Cursor <= maxIdx {
			r := filtered[m.Cursor]
			m.Rows[r.origIndex].Action = "pulling..."
			return m, pullRepoCmd(r.row.Repo, r.origIndex, m.RepoUtils)
		}

	case "P":
		// Pull ALL visible repos.
		var cmds []tea.Cmd
		for _, r := range filtered {
			m.Rows[r.origIndex].Action = "pulling..."
			cmds = append(cmds, pullRepoCmd(r.row.Repo, r.origIndex, m.RepoUtils))
		}
		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}

	case "e":
		// Open exec prompt.
		m.ExecActive = true
		m.ExecInput.Focus()
		return m, textinput.Blink

	case "enter":
		// Open sub-shell in selected repo.
		if len(filtered) > 0 && m.Cursor <= maxIdx {
			r := filtered[m.Cursor]
			fullPath := m.RepoUtils.FullPathWithRepo(r.row.Repo.Path, r.row.Repo.Name)
			return m, openShellCmd(fullPath)
		}

	case "r":
		// Manual refresh.
		if !m.Refreshing {
			return m, m.startRefresh()
		}

	case "/":
		// Open filter.
		m.Filtering = true
		m.FilterInput.Focus()
		return m, textinput.Blink

	case "d":
		// Switch to Discovery view.
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
		// Toggle selection.
		m.Discovery.Selected[m.Discovery.Cursor] = !m.Discovery.Selected[m.Discovery.Cursor]
		// Move cursor down for convenience.
		if m.Discovery.Cursor < maxIdx {
			m.Discovery.Cursor++
		}
	case "enter":
		// Clone all selected repos.
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
		return m, cloneBatchCmd(toClone, m.Config, m.RepoUtils)
	case "q":
		return m, tea.Quit
	}

	return m, nil
}

func truncate(s string, maxLen int) string {
	// Take only the first line.
	if idx := strings.IndexByte(s, '\n'); idx != -1 {
		s = s[:idx]
	}
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
