package tui

import (
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
)

type Styles struct {
	Bold    lipgloss.Style
	Regular lipgloss.Style
	Faint   lipgloss.Style
	Italic  lipgloss.Style
}

var (
	theme       *tint.Registry
	accentColor lipgloss.TerminalColor
	styles      Styles
)

func init() {
	if lipgloss.HasDarkBackground() {
		theme = tint.NewRegistry(tint.TintTomorrowNight)
	} else {
		theme = tint.NewRegistry(tint.TintTomorrow)
	}
	accentColor = theme.BrightPurple()

	styles = Styles{
		Bold:    lipgloss.NewStyle().Foreground(theme.Fg()).Bold(true),
		Regular: lipgloss.NewStyle().Foreground(theme.Fg()),
		Faint:   lipgloss.NewStyle().Foreground(theme.Fg()).Faint(true),
		Italic:  lipgloss.NewStyle().Foreground(theme.Fg()).Italic(true),
	}
}
