package utils

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Copy-pasted from github.com/muesli/termenv@v0.9.0/termenv_unix.go.
// TODO: Refactor after, [feature](https://ï.at/stderr) implemented.
func colorProfile() termenv.Profile {
	term := os.Getenv("TERM")
	colorTerm := os.Getenv("COLORTERM")

	switch strings.ToLower(colorTerm) {
	case "24bit":
		fallthrough
	case "truecolor":
		if term == "screen" || !strings.HasPrefix(term, "screen") {
			// enable TrueColor in tmux, but not for old-school screen
			return termenv.TrueColor
		}
	case "yes":
		fallthrough
	case "true":
		return termenv.ANSI256
	}

	if strings.Contains(term, "256color") {
		return termenv.ANSI256
	}
	if strings.Contains(term, "color") {
		return termenv.ANSI
	}

	return termenv.Ascii
}

var NewErrorCmd = func(format string, values any) func() tea.Msg {
	return SendMsg(fmt.Errorf(format, values))
}

func SendMsg[T any](msg T) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
