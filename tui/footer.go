package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/utils"
)

type Footer struct {
	help.Model
	KeyMap KeyMap
}

func NewFooter() Footer {
	keymap := KeyMap{}
	m := help.New()
	m.Styles.ShortKey = DefaultStyles.Primary
	m.Styles.ShortDesc = DefaultStyles.Primary
	m.Styles.ShortSeparator = DefaultStyles.Primary

	return Footer{
		Model:  m,
		KeyMap: keymap,
	}
}

type KeyMap struct {
	Actions []Action
}

func (k KeyMap) ShortHelp() []key.Binding {
	bindings := make([]key.Binding, len(k.Actions))
	for i, action := range k.Actions {
		prettyKey := strings.ReplaceAll(action.Shortcut, "enter", "↵")
		prettyKey = strings.ReplaceAll(prettyKey, "ctrl", "⌃")
		bindings[i] = key.NewBinding(key.WithKeys(prettyKey), key.WithHelp(prettyKey, action.Title))
	}
	return bindings
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return nil
}

func (f Footer) Update(msg tea.Msg) (Footer, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		for _, action := range f.KeyMap.Actions {
			if action.Shortcut == msg.String() {
				return f, action.SendMsg
			}
		}
	}
	return f, nil
}

func (f *Footer) SetActions(actions ...Action) {
	for i := range actions {
		if actions[i].Shortcut != "" {
			continue
		}

		if i == 0 {
			actions[i].Shortcut = "enter"
		} else if i < 10 {
			actions[i].Shortcut = fmt.Sprintf("ctrl+%d", i)
		}
	}
	f.KeyMap = KeyMap{Actions: actions}
}

func (f Footer) View() string {
	horizontal := strings.Repeat("─", f.Width)
	helpView := f.Model.View(f.KeyMap)
	helpWidth := lipgloss.Width(helpView)
	blanks := strings.Repeat(" ", utils.Max(f.Width-helpWidth-2, 0))

	helpView = lipgloss.JoinHorizontal(lipgloss.Left, blanks, helpView)
	helpView = lipgloss.NewStyle().Padding(0, 1).Render(helpView)
	return lipgloss.JoinVertical(lipgloss.Left, horizontal, helpView)
}
