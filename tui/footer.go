package tui

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type Footer struct {
	help.Model
	KeyMap KeyMap
}

func NewFooter(actions ...Action) *Footer {
	keymap := KeyMap{actions: actions}
	return &Footer{
		Model:  help.New(),
		KeyMap: keymap,
	}
}

type KeyMap struct {
	actions []Action
}

func (k KeyMap) ShortHelp() []key.Binding {
	bindings := make([]key.Binding, len(k.actions))
	for i, action := range k.actions {
		if i == 0 {
			bindings[i] = key.NewBinding(key.WithKeys("↩"), key.WithHelp("↩", action.Title()))
			continue
		}
		bindings[i] = key.NewBinding(key.WithKeys(action.Keybind()), key.WithHelp(action.Keybind(), action.Title()))
	}
	return bindings
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return nil
}

func (f *Footer) SetActions(actions []Action) {
	log.Println("set actions", actions)
	f.KeyMap = KeyMap{actions: actions}
}

func (f Footer) View() string {
	horizontal := strings.Repeat("─", f.Width)
	return lipgloss.JoinVertical(lipgloss.Left, horizontal, f.Model.View(f.KeyMap))
}
