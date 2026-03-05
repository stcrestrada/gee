package tui

import (
	"fmt"
	"strings"

	"gee/pkg/ui"

	"charm.land/lipgloss/v2"
)

var (
	styleHeader    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	styleDim       = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleTableHead = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Bold(true)
	styleCursor    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	styleAction    = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Italic(true)
	stylePrivate   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleSelected  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	styleStale     = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)
	stylePinned    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleHelpBar   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("238"))
)

// View renders the current screen.
func (m AppModel) View() string {
	switch m.ActiveView {
	case ViewDiscovery:
		return m.viewDiscovery()
	default:
		return m.viewDashboard()
	}
}

func (m AppModel) viewDashboard() string {
	var b strings.Builder

	// --- Header ---
	pinnedCount := 0
	for _, r := range m.Rows {
		if r.Pinned {
			pinnedCount++
		}
	}
	title := fmt.Sprintf(" gee — %d repos", len(m.Rows))
	if pinnedCount > 0 {
		title += fmt.Sprintf(" (%d pinned)", pinnedCount)
	}
	header := styleHeader.Render(title)
	if m.Scanning {
		header += styleDim.Render("  ⟳ scanning...")
	} else if m.Refreshing {
		header += styleDim.Render("  ⟳ refreshing...")
	}
	b.WriteString(header + "\n\n")

	// --- Filter bar ---
	if m.Filtering {
		b.WriteString("  / " + m.FilterInput.View() + "\n\n")
	} else if m.Filter != "" {
		b.WriteString(styleDim.Render(fmt.Sprintf("  filter: %s  (/ to edit, esc to clear)", m.Filter)) + "\n\n")
	}

	// --- Table header ---
	headerLine := fmt.Sprintf("  %-2s %-2s %-20s %-15s %-12s %s", "", "", "REPO", "BRANCH", "SYNC", "CHANGES")
	b.WriteString(styleTableHead.Render(headerLine) + "\n")

	// --- Repo rows ---
	filtered := m.filteredRows()

	// Scrolling: compute visible window based on terminal height.
	// Reserve lines for: header(2) + filter(2 max) + table header(1) + log(4 max) + help(2) + exec(2 max) = ~13 overhead
	overhead := 13
	if m.Filter != "" || m.Filtering {
		overhead += 2
	}
	visibleRows := m.Height - overhead
	if visibleRows < 5 {
		visibleRows = 5
	}
	if visibleRows > len(filtered) {
		visibleRows = len(filtered)
	}

	scrollOffset := 0
	if m.Cursor >= visibleRows {
		scrollOffset = m.Cursor - visibleRows + 1
	}
	endIdx := scrollOffset + visibleRows
	if endIdx > len(filtered) {
		endIdx = len(filtered)
	}

	for i := scrollOffset; i < endIdx; i++ {
		fr := filtered[i]
		row := fr.row
		selected := i == m.Cursor

		line := renderDashboardRow(row, selected)
		b.WriteString(line + "\n")
	}

	if len(filtered) == 0 {
		if m.Filter != "" {
			b.WriteString(styleDim.Render("  no repos match filter") + "\n")
		} else if m.Scanning {
			b.WriteString(styleDim.Render("  no repos found — scanning...") + "\n")
		} else {
			b.WriteString(styleDim.Render("  no repos found — run gee add in a git repo to pin it") + "\n")
		}
	}

	// Scroll indicator
	if len(filtered) > visibleRows {
		b.WriteString(styleDim.Render(fmt.Sprintf("  (%d/%d)", m.Cursor+1, len(filtered))) + "\n")
	}

	// --- Action log (last 3 entries) ---
	if len(m.ActionLog) > 0 {
		b.WriteString("\n")
		start := len(m.ActionLog) - 3
		if start < 0 {
			start = 0
		}
		for _, entry := range m.ActionLog[start:] {
			b.WriteString(styleDim.Render("  "+entry) + "\n")
		}
	}

	// --- Exec input ---
	if m.ExecActive {
		b.WriteString("\n  exec> " + m.ExecInput.View() + "\n")
	}

	// --- Help bar ---
	b.WriteString("\n" + m.renderHelpBar())

	return b.String()
}

func renderDashboardRow(row RepoRow, selected bool) string {
	var parts []string

	// Cursor indicator
	cursor := "  "
	if selected {
		cursor = styleCursor.Render("▸ ")
	}

	// Pin icon
	pin := "  "
	if row.Pinned {
		pin = stylePinned.Render("* ")
	}

	// Status icon
	var icon string
	switch {
	case row.Loading:
		icon = styleDim.Render("⠋")
	case row.Failed:
		icon = ui.SymbolError()
	default:
		icon = ui.SymbolSuccess()
	}

	// Repo name
	name := ui.StyleRepoName.Render(fmt.Sprintf("%-20s", row.Repo.Name))

	// Action indicator (inline)
	if row.Action != "" {
		parts = append(parts, cursor+pin+icon+"  "+name+"  "+styleAction.Render(row.Action))
		return strings.Join(parts, "")
	}

	if row.Loading || row.Failed {
		status := "loading..."
		if row.Failed {
			status = "failed"
		}
		parts = append(parts, cursor+pin+icon+"  "+name+"  "+styleDim.Render(status))
		return strings.Join(parts, "")
	}

	s := row.Status

	// Branch with state annotation
	branchDisplay := s.Branch
	if s.State != "" {
		if s.Progress != "" {
			branchDisplay = fmt.Sprintf("%s|%s %s", s.Branch, s.State, s.Progress)
		} else {
			branchDisplay = fmt.Sprintf("%s|%s", s.Branch, s.State)
		}
		branchDisplay = ui.StyleWarning.Render(fmt.Sprintf("%-15s", branchDisplay))
	} else {
		branchDisplay = ui.StyleCommand.Render(fmt.Sprintf("%-15s", branchDisplay))
	}

	// Ahead/behind
	syncParts := []string{}
	if s.Ahead > 0 {
		syncParts = append(syncParts, ui.StyleSuccess.Render(fmt.Sprintf("↑%d", s.Ahead)))
	}
	if s.Behind > 0 {
		syncParts = append(syncParts, ui.StyleError.Render(fmt.Sprintf("↓%d", s.Behind)))
	}
	syncDisplay := strings.Join(syncParts, " ")
	if syncDisplay == "" {
		syncDisplay = styleDim.Render("—")
	}

	// Changes
	var changes []string
	if s.Conflicts > 0 {
		changes = append(changes, ui.StyleError.Render(fmt.Sprintf("!%d conflict", s.Conflicts)))
	}
	if s.Staged > 0 {
		changes = append(changes, ui.StyleSuccess.Render(fmt.Sprintf("+%d staged", s.Staged)))
	}
	if s.Modified > 0 {
		changes = append(changes, ui.StyleWarning.Render(fmt.Sprintf("~%d modified", s.Modified)))
	}
	if s.Untracked > 0 {
		changes = append(changes, ui.StyleCommand.Render(fmt.Sprintf("?%d untracked", s.Untracked)))
	}

	changeDisplay := ui.StyleSuccess.Render("clean")
	if len(changes) > 0 {
		changeDisplay = strings.Join(changes, " ")
	}

	// Staleness badge
	staleDisplay := ""
	if s.Stale {
		staleDisplay = " " + styleStale.Render("STALE")
	}

	return fmt.Sprintf("%s%s%s  %s  %s  %-12s  %s%s", cursor, pin, icon, name, branchDisplay, syncDisplay, changeDisplay, staleDisplay)
}

func (m AppModel) viewDiscovery() string {
	var b strings.Builder

	b.WriteString(styleHeader.Render(fmt.Sprintf(" Discovery (%s)", m.Discovery.Provider)) + "\n\n")

	if m.Discovery.Loading {
		b.WriteString(styleDim.Render("  Loading remote repos...") + "\n")
		b.WriteString("\n" + styleHelpBar.Render("  esc:back  q:quit"))
		return b.String()
	}

	if m.Discovery.Error != nil {
		b.WriteString(ui.StyleError.Render(fmt.Sprintf("  Error: %s", m.Discovery.Error)) + "\n")
		b.WriteString("\n" + styleHelpBar.Render("  esc:back  q:quit"))
		return b.String()
	}

	if len(m.Discovery.RemoteRepos) == 0 {
		b.WriteString(styleDim.Render("  No remote repos found.") + "\n")
		b.WriteString("\n" + styleHelpBar.Render("  esc:back  q:quit"))
		return b.String()
	}

	// Count selections
	selCount := 0
	for _, sel := range m.Discovery.Selected {
		if sel {
			selCount++
		}
	}

	// Table header
	headerLine := fmt.Sprintf("  %-2s %-3s %-40s %s", "", "SEL", "REPOSITORY", "DESCRIPTION")
	b.WriteString(styleTableHead.Render(headerLine) + "\n")

	// Scrolling
	overhead := 10
	visibleRows := m.Height - overhead
	if visibleRows < 5 {
		visibleRows = 5
	}
	if visibleRows > len(m.Discovery.RemoteRepos) {
		visibleRows = len(m.Discovery.RemoteRepos)
	}

	scrollOffset := 0
	if m.Discovery.Cursor >= visibleRows {
		scrollOffset = m.Discovery.Cursor - visibleRows + 1
	}
	endIdx := scrollOffset + visibleRows
	if endIdx > len(m.Discovery.RemoteRepos) {
		endIdx = len(m.Discovery.RemoteRepos)
	}

	for i := scrollOffset; i < endIdx; i++ {
		repo := m.Discovery.RemoteRepos[i]
		cursor := "  "
		if i == m.Discovery.Cursor {
			cursor = styleCursor.Render("▸ ")
		}

		sel := styleDim.Render("[ ]")
		if m.Discovery.Selected[i] {
			sel = styleSelected.Render("[✓]")
		}

		nameStr := repo.FullName
		if repo.Private {
			nameStr += " " + stylePrivate.Render("(private)")
		}
		name := fmt.Sprintf("%-40s", nameStr)

		desc := repo.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}

		b.WriteString(fmt.Sprintf("%s%s  %s  %s\n", cursor, sel, name, styleDim.Render(desc)))
	}

	// Scroll indicator
	if len(m.Discovery.RemoteRepos) > visibleRows {
		b.WriteString(styleDim.Render(fmt.Sprintf("  (%d/%d)", m.Discovery.Cursor+1, len(m.Discovery.RemoteRepos))) + "\n")
	}

	if selCount > 0 {
		b.WriteString("\n" + ui.StyleSuccess.Render(fmt.Sprintf("  %d selected", selCount)) + "\n")
	}

	// Help bar
	b.WriteString("\n" + styleHelpBar.Render("  j/k:nav  space:select  enter:clone selected  esc:back  q:quit"))

	return b.String()
}

func (m AppModel) renderHelpBar() string {
	keys := []string{"j/k:nav", "a:pin", "p:pull", "P:pull all", "e:exec", "↵:cd", "r:refresh", "/:filter"}
	if m.Discovery.Provider != "" {
		keys = append(keys, "d:discover")
	}
	keys = append(keys, "q:quit")
	return styleHelpBar.Render("  " + strings.Join(keys, "  "))
}
