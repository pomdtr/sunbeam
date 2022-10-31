package tui

import (
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Selection lipgloss.Style
	Primary   lipgloss.Style
	Secondary lipgloss.Style
	Focused   lipgloss.Style
}

var DefaultStyles Styles = Styles{
	Selection: lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
	Primary:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}),
	Secondary: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9a9a9c", Dark: "#777777"}),
	Focused:   lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{Light: "255", Dark: "0"}),
}
