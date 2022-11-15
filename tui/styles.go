package tui

import (
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
)

type Styles struct {
	Bold    lipgloss.Style
	Regular lipgloss.Style
	Faint   lipgloss.Style
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
		Bold:    lipgloss.NewStyle().Background(theme.Bg()).Foreground(theme.Fg()).Bold(true),
		Regular: lipgloss.NewStyle().Background(theme.Bg()).Foreground(theme.Fg()),
		Faint:   lipgloss.NewStyle().Background(theme.Bg()).Foreground(theme.Fg()).Faint(true),
	}
}
