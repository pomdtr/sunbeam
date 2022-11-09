package tui

import (
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
)

type Styles struct {
	Accent     lipgloss.Style
	Primary    lipgloss.Style
	Secondary  lipgloss.Style
	Background lipgloss.Style
}

var (
	theme       *tint.Registry
	accentColor lipgloss.TerminalColor
	styles      Styles
)

func init() {
	if lipgloss.HasDarkBackground() {
		theme = tint.NewRegistry(tint.TintBuiltinSolarizedDark)
	} else {
		theme = tint.NewRegistry(tint.TintBuiltinSolarizedLight)
	}
	accentColor = theme.BrightPurple()

	styles = Styles{
		Accent:     lipgloss.NewStyle().Background(theme.Bg()).Foreground(accentColor),
		Primary:    lipgloss.NewStyle().Background(theme.Bg()).Foreground(theme.Fg()).Bold(true),
		Secondary:  lipgloss.NewStyle().Background(theme.Bg()).Foreground(theme.Fg()),
		Background: lipgloss.NewStyle().Background(theme.Bg()),
	}
}
