package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/commands"
)

func Start() error {
	rootItems := make([]commands.ListItem, len(commands.Commands))

	for i, command := range commands.Commands {
		rootItems[i] = commands.ListItem{
			Title:    command.Title,
			Subtitle: command.Subtitle,
			Actions: []commands.ScriptAction{
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
