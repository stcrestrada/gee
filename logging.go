package main

import (
	"fmt"
	"github.com/fatih/color"
)

func Info(format string, args ...interface{}) {
	c := color.New(color.FgGreen, color.Bold)
	c.Printf("%s\n", fmt.Sprintf(format, args...))
}

// CheckIfError should be used to naively panic if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}
	c := color.New(color.FgHiRed, color.Bold)
	c.Printf("%s\n", fmt.Sprintf("error: %s", err))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	c := color.New(color.FgYellow, color.Bold)
	c.Printf("%s\n", fmt.Sprintf(format, args...))
}

func WarningRed(format string, args ...interface{}) {
	c := color.New(color.FgRed, color.Bold)
	c.Printf("%s\n", fmt.Sprintf(format, args...))
}
