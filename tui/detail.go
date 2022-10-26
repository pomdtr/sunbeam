package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Detail struct {
	title         string
	width, height int

	viewport viewport.Model
}

func NewDetail(title string, content string) *Detail {
	viewport := viewport.New(0, 0)
	viewport.SetContent(content)
	return &Detail{title: title, viewport: viewport}
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
	c.viewport.Height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footerView())
}

func (c *Detail) View() string {
	log.Println(c.viewport.View())
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.viewport.View(), c.footerView())
}

func (c *Detail) SetContent(content string) {
	c.viewport.SetContent(content)
}

func (c *Detail) headerView() string {
	return SunbeamHeader(c.width)
}

func (c *Detail) footerView() string {
	return SunbeamFooter(c.width, c.title)
}
