package util

import (
	"charm.land/lipgloss/v2"
	"fmt"
)

var Verbose bool

// SetVerbose sets the logging verbosity
func SetVerbose(verbose bool) {
	Verbose = verbose
}

// WarningError represents a custom warning error with a message
type WarningError struct {
	Message string
}

func (w *WarningError) Error() string {
	return w.Message
}

// NewWarning creates a new WarningError
func NewWarning(message string) error {
	return &WarningError{Message: message}
}

// InfoError represents a custom info error with a message
type InfoError struct {
	Message string
}

func (i *InfoError) Error() string {
	return i.Message
}

// NewInfo creates a new InfoError
func NewInfo(message string) error {
	return &InfoError{
		Message: message,
	}
}

var (
	styleSymbolGreen = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	styleSymbolRed   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	styleSymbolYellow = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))
	styleSymbolCyan  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	styleMessage     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
)

// Info logs an info message with a green check mark and white message
func Info(format string, args ...interface{}) {
	fmt.Printf("%s %s\n", styleSymbolGreen.Render("✓"), styleMessage.Render(fmt.Sprintf(format, args...)))
}

// VerboseLog logs a verbose message if verbose logging is enabled
func VerboseLog(format string, args ...interface{}) {
	if Verbose {
		fmt.Printf("%s %s\n", styleSymbolCyan.Render("✓"), styleMessage.Render(fmt.Sprintf(format, args...)))
	}
}

// CheckIfError logs an error message with a red cross and panics if the error is not nil
func CheckIfError(err error) {
	if err == nil {
		return
	}
	fmt.Printf("%s %s\n", styleSymbolRed.Render("✗"), styleMessage.Render(fmt.Sprintf("error: %s", err)))
}

// Warning logs a warning message with a yellow exclamation mark and white message
func Warning(format string, args ...interface{}) {
	fmt.Printf("%s %s\n", styleSymbolYellow.Render("!"), styleMessage.Render(fmt.Sprintf(format, args...)))
}

// WarningRed logs a warning message with a red exclamation mark and white message
func WarningRed(format string, args ...interface{}) {
	fmt.Printf("%s %s\n", styleSymbolRed.Render("✗"), styleMessage.Render(fmt.Sprintf(format, args...)))
}
