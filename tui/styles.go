package tui

import (
	"os"

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
	lightTint := tint.TintTomorrowNight
	darkTint := tint.TintTomorrow
	switch os.Getenv("SUNBEAM_APPEARANCE") {
	case "dark":
		lipgloss.SetHasDarkBackground(true)
		theme = tint.NewRegistry(lightTint)
	case "light":
		lipgloss.SetHasDarkBackground(false)
		theme = tint.NewRegistry(darkTint)
	case "auto", "":
		// lipgloss default detection
		if lipgloss.HasDarkBackground() {
			theme = tint.NewRegistry(lightTint)
		} else {
			theme = tint.NewRegistry(darkTint)
		}
	}

	accentColor = theme.BrightPurple()

	styles = Styles{
		Bold:    lipgloss.NewStyle().Foreground(theme.Fg()).Bold(true),
		Regular: lipgloss.NewStyle().Foreground(theme.Fg()),
		Faint:   lipgloss.NewStyle().Foreground(theme.Fg()).Faint(true),
		Italic:  lipgloss.NewStyle().Foreground(theme.Fg()).Italic(true),
	}
}
