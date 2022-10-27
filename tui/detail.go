package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Detail struct {
	width, height int

	viewport viewport.Model
	actions  []Action
	footer   *Footer
}

func NewDetail(content string, actions []Action) *Detail {
	viewport := viewport.New(0, 0)
	viewport.SetContent(content)
	footer := NewFooter(actions...)
	return &Detail{viewport: viewport, actions: actions, footer: footer}
}

func (c *Detail) Init() tea.Cmd {
	return c.viewport.Init()
}

func (c *Detail) Update(msg tea.Msg) (*Detail, tea.Cmd) {
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
		default:
			for _, action := range c.actions {
				if action.Keybind() == msg.String() {
					return c, action.Exec()
				}
			}

		}

	}
	var cmd tea.Cmd
	c.viewport, cmd = c.viewport.Update(msg)
	return c, cmd
}

func (c *Detail) SetSize(width, height int) {
	c.height = height
	c.width = width
	c.viewport.Width = width
	c.viewport.Height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footer.View())
}

func (c *Detail) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.viewport.View(), c.footer.View())
}

func (c *Detail) SetContent(content string) {
	c.viewport.SetContent(content)
}

func (c *Detail) headerView() string {
	return SunbeamHeader(c.width)
}
