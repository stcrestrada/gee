package util

import (
	"fmt"
	"github.com/fatih/color"
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

// logMessage is a helper function to print messages with specific color attributes and symbols
func logMessage(symbolColor color.Attribute, symbol string, messageColor color.Attribute, format string, args ...interface{}) {
	symbolColored := color.New(symbolColor, color.Bold).Sprintf(symbol)
	messageColored := color.New(messageColor, color.Bold).Sprintf(format, args...)
	fmt.Printf("%s %s\n", symbolColored, messageColored)
}

// Info logs an info message with a green check mark and white message
func Info(format string, args ...interface{}) {
	logMessage(color.FgGreen, "✓", color.FgWhite, format, args...)
}

// VerboseLog logs a verbose message if verbose logging is enabled
func VerboseLog(format string, args ...interface{}) {
	if Verbose {
		logMessage(color.FgCyan, "✓", color.FgWhite, format, args...)
	}
}

// CheckIfError logs an error message with a red cross and panics if the error is not nil
func CheckIfError(err error) {
	if err == nil {
		return
	}
	logMessage(color.FgHiRed, "✗", color.FgWhite, "error: %s", err)
}

// Warning logs a warning message with a yellow exclamation mark and white message
func Warning(format string, args ...interface{}) {
	logMessage(color.FgYellow, "!", color.FgWhite, format, args...)
}

// WarningRed logs a warning message with a red exclamation mark and white message
func WarningRed(format string, args ...interface{}) {
	logMessage(color.FgRed, "✗", color.FgWhite, format, args...)
}
