package bubbles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func SunbeamHeader(width int) string {
	header := strings.Repeat(" ", width)
	separator := strings.Repeat("─", width)
	return lipgloss.JoinVertical(lipgloss.Left, header, separator)
}
