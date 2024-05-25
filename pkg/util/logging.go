package util

import (
	"fmt"
	"github.com/fatih/color"
)

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

// logMessage is a helper function to print messages with specific color attributes
func logMessage(colorAttribute color.Attribute, format string, args ...interface{}) {
	c := color.New(colorAttribute, color.Bold)
	c.Printf("%s\n", fmt.Sprintf(format, args...))
}

// Info logs an info message with green text and bold style
func Info(format string, args ...interface{}) {
	logMessage(color.FgGreen, format, args...)
}

// CheckIfError logs an error message and panics if the error is not nil
func CheckIfError(err error) {
	if err == nil {
		return
	}
	logMessage(color.FgHiRed, "error: %s", err)
	panic(err)
}

// Warning logs a warning message with yellow text and bold style
func Warning(format string, args ...interface{}) {
	logMessage(color.FgYellow, format, args...)
}

// WarningRed logs a warning message with red text and bold style
func WarningRed(format string, args ...interface{}) {
	logMessage(color.FgRed, format, args...)
}
