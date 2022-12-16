package tui

import (
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Bold   lipgloss.Style
	Faint  lipgloss.Style
	Italic lipgloss.Style
}

var (
	accentColor lipgloss.TerminalColor
	styles      Styles
)

func init() {
	accentColor = lipgloss.Color("13")

	styles = Styles{
		Bold:   lipgloss.NewStyle().Bold(true),
		Faint:  lipgloss.NewStyle().Faint(true),
		Italic: lipgloss.NewStyle().Italic(true),
	}
}
