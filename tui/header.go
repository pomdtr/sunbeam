package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func SunBeamHeaderWithInput(width int, t *textinput.Model) string {
	header := t.View()
	separator := strings.Repeat("─", width)
	return lipgloss.JoinVertical(lipgloss.Left, header, separator)
}

func SunbeamHeader(width int) string {
	header := strings.Repeat(" ", width)
	separator := strings.Repeat("─", width)
	return lipgloss.JoinVertical(lipgloss.Left, header, separator)
}
