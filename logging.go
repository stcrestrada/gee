package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	c := color.New(color.FgYellow, color.Bold)
	c.Printf("%s \n", fmt.Sprintf(format, args...))
}
