package pages

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/bubbles"
	"github.com/pomdtr/sunbeam/scripts"
)

type ActionRunner func(scripts.ScriptAction) tea.Cmd

type DetailContainer struct {
	response  scripts.DetailResponse
	runAction ActionRunner
	width     int
	height    int
	viewport  *viewport.Model
}

func NewDetailContainer(response *scripts.DetailResponse) *DetailContainer {
	viewport := viewport.New(0, 0)
	viewport.SetContent(response.Text)

	return &DetailContainer{
		response: *response,
		viewport: &viewport,
	}
}

func (c *DetailContainer) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.viewport.Width = width
	c.viewport.Height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footerView())
}

func (c *DetailContainer) Init() tea.Cmd {
	return nil
}

func (c *DetailContainer) headerView() string {
	return bubbles.SunbeamHeader(c.viewport.Width)
}

func (c *DetailContainer) footerView() string {
	return bubbles.SunbeamFooter(c.viewport.Width, c.response.Title)
}

func (c *DetailContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			for _, action := range c.response.Actions {
				if action.Keybind == string(msg.Runes) {
					return c, c.runAction(action)
				}
			}
		case tea.KeyEscape:
			return c, PopCmd
		}
	}

	var cmd tea.Cmd
	model, cmd := c.viewport.Update(msg)
	c.viewport = &model
	return c, cmd
}

func (c *DetailContainer) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.viewport.View(), c.footerView())
}
