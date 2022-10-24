package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/sunbeam"
)

func Draw() error {
	rootItems := make([]sunbeam.ListItem, len(sunbeam.Commands))

	for i, command := range sunbeam.Commands {
		rootItems[i] = sunbeam.ListItem{
			Title:    command.Title,
			Subtitle: command.Subtitle,
			Actions: []sunbeam.ScriptAction{
				{Type: "push", Target: command.Id, Root: command.Root.String()},
			},
		}
	}

	rootContainer := NewListContainer("Commands", rootItems, RunAction)
	m := NewRootModel(rootContainer)
	p := tea.NewProgram(m, tea.WithAltScreen())
	err := p.Start()

	if err != nil {
		return err
	}
	return nil
}
