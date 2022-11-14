package tui

import (
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
)

type Styles struct {
	Title      lipgloss.Style
	Text       lipgloss.Style
	Background lipgloss.Style
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
		Title:      lipgloss.NewStyle().Background(theme.Bg()).Foreground(theme.Fg()).Bold(true),
		Text:       lipgloss.NewStyle().Background(theme.Bg()).Foreground(theme.Fg()),
		Background: lipgloss.NewStyle().Background(theme.Bg()),
	}
}
