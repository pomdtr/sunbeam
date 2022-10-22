package cli

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DetailContainer struct {
	width, height int
	title         string
	viewport      viewport.Model
}

func NewDetailContainer(title string, content string) *DetailContainer {
	viewport := viewport.New(0, 0)
	viewport.SetContent(content)
	return &DetailContainer{viewport: viewport, title: title}
}

func (c *DetailContainer) Init() tea.Cmd {
	return nil
}

func (c *DetailContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			switch msg.String() {
			case "q", "Q":
				return nil, tea.Quit
			}
		case tea.KeyEscape:
			return c, PopCmd
		}
	}
	var cmd tea.Cmd
	c.viewport, cmd = c.viewport.Update(msg)
	return c, cmd
}

func (c *DetailContainer) headerView() string {
	return SunbeamHeader(c.width)
}

func (c *DetailContainer) footerView() string {
	return SunbeamFooter(c.width, c.title)
}

func (c *DetailContainer) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.viewport.Height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footerView())
}

func (c *DetailContainer) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.viewport.View(), c.footerView())
}
