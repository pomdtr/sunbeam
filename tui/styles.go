package tui

import (
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
)

type Colors struct {
	Accent     lipgloss.TerminalColor
	Primary    lipgloss.TerminalColor
	Secondary  lipgloss.TerminalColor
	Background lipgloss.TerminalColor
}

type Styles struct {
	Accent     lipgloss.Style
	Primary    lipgloss.Style
	Secondary  lipgloss.Style
	Background lipgloss.Style
}

var (
	theme  *tint.Registry
	colors Colors
	styles Styles
	blank  string
)

func init() {
	theme = tint.NewRegistry(tint.TintBuiltinSolarizedDark)
	colors = Colors{
		Accent:     theme.Purple(),
		Primary:    theme.Fg(),
		Secondary:  theme.Fg(),
		Background: theme.Bg(),
	}
	styles = Styles{
		Accent:     lipgloss.NewStyle().Background(colors.Background).Foreground(colors.Accent),
		Primary:    lipgloss.NewStyle().Background(colors.Background).Foreground(colors.Primary).Bold(true),
		Secondary:  lipgloss.NewStyle().Background(colors.Background).Foreground(colors.Secondary),
		Background: lipgloss.NewStyle().Background(colors.Background),
	}
	blank = styles.Background.Render(" ")
}
