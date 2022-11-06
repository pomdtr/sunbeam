package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/utils"
)

type Footer struct {
	title string
	help.Model
	bindings []key.Binding
}

func NewFooter(title string) Footer {
	m := help.New()
	m.Styles.ShortKey = DefaultStyles.Primary
	m.Styles.ShortDesc = DefaultStyles.Primary
	m.Styles.ShortSeparator = DefaultStyles.Primary

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

	if len(f.bindings) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, horizontal, fmt.Sprintf(" %s ", f.title))
	}

	shortHelp := f.Model.ShortHelpView(f.bindings)
	blanks := strings.Repeat(" ", utils.Max(f.Width-lipgloss.Width(shortHelp)-len(f.title)-3, 0))

	shortHelp = lipgloss.JoinHorizontal(lipgloss.Left, " ", f.title, " ", blanks, shortHelp, " ")

	return lipgloss.JoinVertical(lipgloss.Left, horizontal, shortHelp)
}
