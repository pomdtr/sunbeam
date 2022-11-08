package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type Footer struct {
	title string
	help.Model
	bindings []key.Binding
}

func NewFooter(title string) Footer {
	m := help.New()
	m.Styles.ShortKey = styles.Primary
	m.Styles.ShortDesc = styles.Primary
	m.Styles.ShortSeparator = styles.Secondary

	return Footer{
		Model: m,
		title: title,
	}
}

func (f *Footer) SetBindings(bindings ...key.Binding) {
	f.bindings = bindings
}

func (f Footer) View() string {
	horizontal := strings.Repeat("─", f.Width)
	horizontal = styles.Primary.Render(horizontal)

	if len(f.bindings) == 0 {
		title := styles.Primary.Copy().Padding(0, 1).Width(f.Width).Render(f.title)
		return lipgloss.JoinVertical(lipgloss.Left, horizontal, title)
	}

	title := styles.Primary.Copy().Padding(0, 1).Render(f.title)
	shortHelp := f.Model.ShortHelpView(f.bindings)

	availableWidth := f.Width - lipgloss.Width(title)
	shortHelp = styles.Background.Copy().Padding(0, 1).Width(availableWidth).Align(lipgloss.Right).Render(shortHelp)

	shortHelp = lipgloss.JoinHorizontal(lipgloss.Left, title, shortHelp)

	return lipgloss.JoinVertical(lipgloss.Left, horizontal, shortHelp)
}
