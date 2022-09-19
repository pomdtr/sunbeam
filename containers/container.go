package containers

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Container interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Container, tea.Cmd)
	View() string
	SetSize(width, height int)
}
