package bubbles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/utils"
)

func SunbeamFooter(width int, titleLabel string) string {
	separator := strings.Repeat("─", width)
	title := lipgloss.NewStyle().PaddingRight(1).Render(titleLabel)
	footer := strings.Repeat(" ", utils.Max(0, width-lipgloss.Width(title)))
	footer = lipgloss.JoinHorizontal(lipgloss.Center, title, footer)
	return lipgloss.JoinVertical(lipgloss.Left, separator, footer)
}

func SunbeamFooterWithActions(width int, titleLabel string, activateLabel string) string {
	title := lipgloss.NewStyle().PaddingRight(1).Render(titleLabel)
	activateButton := lipgloss.NewStyle().PaddingLeft(1).Render(fmt.Sprintf("%s ↩", activateLabel))
	separator := lipgloss.NewStyle().Padding(0, 1).Render("|")
	actionsButton := "Actions ^P"
	line := strings.Repeat(" ", utils.Max(0, width-lipgloss.Width(title)-lipgloss.Width(activateButton)-lipgloss.Width(separator)-lipgloss.Width(actionsButton)))
	footer := lipgloss.JoinHorizontal(lipgloss.Center, title, line, activateButton, separator, actionsButton)
	horizontal := strings.Repeat("─", width)

	return lipgloss.JoinVertical(lipgloss.Left, horizontal, footer)
}
