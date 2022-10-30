package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/utils"
)

type Footer struct {
	help.Model
	KeyMap KeyMap
}

func NewFooter(actions ...Action) *Footer {
	keymap := KeyMap{actions: actions}
	m := help.New()
	m.Styles.ShortKey = DefaultStyles.Primary
	m.Styles.ShortDesc = DefaultStyles.Primary
	m.Styles.ShortSeparator = DefaultStyles.Primary

	return &Footer{
		Model:  m,
		KeyMap: keymap,
	}
}

type KeyMap struct {
	actions []Action
}

func (k KeyMap) ShortHelp() []key.Binding {
	bindings := make([]key.Binding, len(k.actions))
	for i, action := range k.actions {
		prettyKey := strings.ReplaceAll(action.Shortcut, "enter", "↵")
		prettyKey = strings.ReplaceAll(prettyKey, "ctrl", "⌃")
		bindings[i] = key.NewBinding(key.WithKeys(prettyKey), key.WithHelp(prettyKey, action.Title))
	}
	return bindings
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return nil
}

func (f *Footer) SetActions(actions []Action) {
	f.KeyMap = KeyMap{actions: actions}
}

func (f Footer) View() string {
	horizontal := strings.Repeat("─", f.Width)
	helpView := f.Model.View(f.KeyMap)
	helpWidth := lipgloss.Width(helpView)
	blanks := strings.Repeat(" ", utils.Max(f.Width-helpWidth, 0))

	helpView = lipgloss.JoinHorizontal(lipgloss.Left, blanks, helpView)
	return lipgloss.JoinVertical(lipgloss.Left, horizontal, helpView)
}
