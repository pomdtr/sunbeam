package list

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the menu menu.
type KeyMap struct {
	// Keybindings used when browsing the list.
	CursorUp   key.Binding
	CursorDown key.Binding
	NextPage   key.Binding
	PrevPage   key.Binding
	GoToStart  key.Binding
	GoToEnd    key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("pgup"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("pgdown"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys("home"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys("end"),
		),
	}
}
