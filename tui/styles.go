package tui

import (
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Selection lipgloss.Style
	Primary   lipgloss.Style
	Secondary lipgloss.Style
}

var DefaultStyles Styles = Styles{
	Selection: lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
	Primary:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}),
	Secondary: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}),
}
