package bubbles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/utils"
)

func SunbeamFooter(width int, titleLabel string) string {
	title := lipgloss.NewStyle().PaddingRight(1).Render(titleLabel)
	line := strings.Repeat("─", utils.Max(0, width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)

}

func SunbeamFooterWithActions(width int, titleLabel string, activateLabel string) string {
	title := lipgloss.NewStyle().PaddingRight(1).Render(titleLabel)
	activateButton := lipgloss.NewStyle().PaddingLeft(1).Render(fmt.Sprintf("%s ↩", activateLabel))
	separator := lipgloss.NewStyle().Padding(0, 1).Render("|")
	actionsButton := "Actions ^K"
	line := strings.Repeat("─", utils.Max(0, width-lipgloss.Width(title)-lipgloss.Width(activateButton)-lipgloss.Width(separator)-lipgloss.Width(actionsButton)))
	footer := lipgloss.JoinHorizontal(lipgloss.Center, title, line, activateButton, separator, actionsButton)

	return footer
}
