package pages

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/bubbles"
)

type ErrorContainer struct {
	width, height int
	viewport      viewport.Model
}

func NewErrorContainer(err error) *ErrorContainer {
	viewport := viewport.New(0, 0)
	viewport.SetContent(err.Error())
	return &ErrorContainer{viewport: viewport}
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

func (c *ErrorContainer) headerView() string {
	return bubbles.SunbeamHeader(c.width)
}

func (c *ErrorContainer) footerView() string {
	return bubbles.SunbeamFooterWithActions(c.width, "Error", "Exit")
}

func (c *ErrorContainer) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.viewport.Height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footerView())
}

func (c *ErrorContainer) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.viewport.View(), c.footerView())
}
