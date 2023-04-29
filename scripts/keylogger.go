package main

import tea "github.com/charmbracelet/bubbletea"

type KeyLogger struct {
	lastKey string
}

func (k KeyLogger) Init() tea.Cmd {
	return nil
}

func (k KeyLogger) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return k, tea.Quit
		default:
			k.lastKey = msg.String()
		}
	}

	return k, nil
}

func (k KeyLogger) View() string {
	return k.lastKey
}

func main() {
	p := tea.NewProgram(&KeyLogger{})
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
