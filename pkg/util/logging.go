package util

import (
	"fmt"
	"github.com/fatih/color"
)

// Info logs an info message with green text and bold style
func Info(format string, args ...interface{}) {
	c := color.New(color.FgGreen, color.Bold)
	c.Printf("%s\n", fmt.Sprintf(format, args...))
}

// CheckIfError logs an error message and panics if the error is not nil
func CheckIfError(err error) {
	if err == nil {
		return
	}
	c := color.New(color.FgHiRed, color.Bold)
	c.Printf("%s\n", fmt.Sprintf("error: %s", err))
	panic(err)
}

// Warning logs a warning message with yellow text and bold style
func Warning(format string, args ...interface{}) {
	c := color.New(color.FgYellow, color.Bold)
	c.Printf("%s\n", fmt.Sprintf(format, args...))
}

// WarningRed logs a warning message with red text and bold style
func WarningRed(format string, args ...interface{}) {
	c := color.New(color.FgRed, color.Bold)
	c.Printf("%s\n", fmt.Sprintf(format, args...))
}
