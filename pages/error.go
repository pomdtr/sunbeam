package pages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ErrorContainer struct {
	width, height int
	err           error
}

func NewErrorContainer(err error) *ErrorContainer {
	return &ErrorContainer{err: err}
}

func (c *ErrorContainer) Init() tea.Cmd {
	return nil
}

func (c *ErrorContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return c, PopCmd
		case tea.KeyEnter:
			return c, tea.Quit
		}
	}
	return c, nil
}

func (c *ErrorContainer) SetSize(width, height int) {
	c.width = width
	c.height = height
}

func (c *ErrorContainer) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.err.Error(), "Press enter to exit")
}
