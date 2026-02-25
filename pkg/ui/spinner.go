package ui

import (
	"fmt"
	"io"
	"time"
)

type State string

type SpinnerState struct {
	State State
	Msg   string
	Err   string
}

const StateLoading = State("loading")
const StateError = State("error")
const StateSuccess = State("success")

var spinnerUnicodeStates = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}

func PrintSpinnerStates(writer io.Writer, states []*SpinnerState) func() {
	i := 0
	for range states {
		writer.Write([]byte("\n"))
	}

	writeStates := func(i int) bool {
		for line := 0; line < len(states); line++ {
			writer.Write([]byte("\033[2K\033[A"))
		}

		shouldContinue := false
		for _, state := range states {
			var formattedMessage string
			spinnerIcon := spinnerUnicodeStates[i%len(spinnerUnicodeStates)]
			if state.State == StateSuccess {
				spinnerIcon = "✓"
				formattedMessage = fmt.Sprintf("%s %s\n", StyleSuccess.Render(spinnerIcon), StyleRepoName.Render(state.Msg))
			} else if state.State == StateError {
				spinnerIcon = "✗"
				formattedMessage = fmt.Sprintf("%s %s\n", StyleError.Render(spinnerIcon), StyleRepoName.Render(state.Msg))
			} else {
				shouldContinue = true
				formattedMessage = fmt.Sprintf("%s %s\n", StyleSummaryLine.Render(spinnerIcon), StyleRepoName.Render(state.Msg))
			}
			writer.Write([]byte(formattedMessage))
		}
		return shouldContinue
	}
	finishChan := make(chan bool)
	finish := func() {
		finishChan <- true
		writeStates(i)
	}
	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * 100):
				shouldContinue := writeStates(i)
				if !shouldContinue {
					return
				}

			case <-finishChan:
				return
			}
			i++
		}
	}()
	return finish
}
