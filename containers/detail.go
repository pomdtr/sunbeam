package containers

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type DetailContainer struct {
	Markdown string
	*viewport.Model
}

func (c DetailContainer) SetSize(width, height int) {
	c.Width = width
	c.Height = height
}

func (c DetailContainer) Init() tea.Cmd {
	return nil
}

func (c DetailContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	var cmd tea.Cmd
	model, cmd := c.Model.Update(msg)
	c.Model = &model
	return c, cmd
}

func (c DetailContainer) View() string {
	return c.Model.View()
}
