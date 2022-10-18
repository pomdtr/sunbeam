package pages

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Container interface {
	Update(msg tea.Msg) (Container, tea.Cmd)
	View() string
	SetSize(width, height int)
}
