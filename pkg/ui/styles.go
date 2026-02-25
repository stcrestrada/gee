package ui

import (
	"charm.land/lipgloss/v2"
)

var (
	// Header styles
	StyleRepoName    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	StyleCommand     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("245"))
	StyleSummaryLine = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	// Status styles
	StyleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	StyleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	StyleWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	// Box for expanded repo output
	StyleRepoBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			PaddingLeft(1).
			PaddingRight(1)

	StyleRepoBoxError = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("1")).
				PaddingLeft(1).
				PaddingRight(1)

	// Content inside boxes
	StyleStdout = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	StyleStderr = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	// Footer
	StyleFooter = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("238")).
			MarginTop(1)
)

func SymbolSuccess() string {
	return StyleSuccess.Render("✓")
}

func SymbolError() string {
	return StyleError.Render("✗")
}

func SymbolWarning() string {
	return StyleWarning.Render("!")
}
