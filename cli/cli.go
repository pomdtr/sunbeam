package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/commands"
)

func Start() error {
	rootContainer := NewListContainer("Commands", commands.RootItems)
	m := NewRootModel(rootContainer)
	p := tea.NewProgram(m, tea.WithAltScreen())
	err := p.Start()

	if err != nil {
		return err
	}
	return nil
}
