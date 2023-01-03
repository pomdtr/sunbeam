package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/sunbeamlauncher/sunbeam/utils"
)

type Footer struct {
	title    string
	Width    int
	bindings []key.Binding
}

func NewFooter(title string) Footer {
	return Footer{
		title: title,
	}
}

func (f *Footer) SetBindings(bindings ...key.Binding) {
	f.bindings = bindings
}

func (f Footer) View() string {
	horizontal := strings.Repeat("─", f.Width)

	if len(f.bindings) == 0 {
		title := styles.Italic.Copy().Padding(0, 1).Width(f.Width).Render(f.title)
		return lipgloss.JoinVertical(lipgloss.Left, horizontal, title)
	}

	keys := make([]string, len(f.bindings))
	for i, binding := range f.bindings {
		keys[i] = fmt.Sprintf("%s %s", binding.Help().Desc, binding.Help().Key)
	}
	help := strings.Join(keys, " · ")
	help = fmt.Sprintf("  %s ", help)

	availableWidth := utils.Max(0, f.Width-lipgloss.Width(help))
	title := fmt.Sprintf(" %s", f.title)

	if availableWidth < lipgloss.Width(title) {
		title = title[:availableWidth]
	}

	blanks := strings.Repeat(" ", utils.Max(0, f.Width-lipgloss.Width(title)-lipgloss.Width(help)))

	footerRow := lipgloss.JoinHorizontal(lipgloss.Left, styles.Italic.Render(title), blanks, help)

	return lipgloss.JoinVertical(lipgloss.Left, horizontal, footerRow)
}
