package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/utils"
)

type Footer struct {
	title    string
	style    lipgloss.Style
	Width    int
	bindings []key.Binding
}

func NewFooter(title string) Footer {
	return Footer{
		style: styles.Regular.Copy(),
		title: title,
	}
}

func (f *Footer) SetBindings(bindings ...key.Binding) {
	f.bindings = bindings
}

func (f Footer) View() string {
	horizontal := strings.Repeat("─", f.Width)
	horizontal = f.style.Render(horizontal)

	if len(f.bindings) == 0 {
		title := styles.Regular.Copy().Padding(0, 1).Width(f.Width).Render(f.title)
		return lipgloss.JoinVertical(lipgloss.Left, horizontal, title)
	}

	// availableWidth := f.Width - lipgloss.Width(title)
	keys := make([]string, len(f.bindings))
	for i, binding := range f.bindings {
		keys[i] = fmt.Sprintf("%s %s", binding.Help().Desc, binding.Help().Key)
	}
	help := strings.Join(keys, " • ")

	titleView := f.style.Copy().Padding(0, 1).Render(f.title)
	availableWidth := utils.Max(0, f.Width-lipgloss.Width(titleView))
	helpView := f.style.Copy().Width(availableWidth).Align(lipgloss.Right).Padding(0, 1).Render(help)

	footerRow := lipgloss.JoinHorizontal(lipgloss.Top, titleView, helpView)

	return lipgloss.JoinVertical(lipgloss.Left, horizontal, footerRow)
}
