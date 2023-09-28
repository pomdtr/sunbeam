package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type ErrorPage struct {
	statusBar StatusBar
	viewport  viewport.Model
	msg       string
}

func NewErrorPage(err error) *ErrorPage {
	viewport := viewport.New(0, 0)
	viewport.Style = lipgloss.NewStyle().Padding(0, 1)
	viewport.SetContent(err.Error())
	page := ErrorPage{
		statusBar: NewStatusBar(),
		viewport:  viewport,
		msg:       err.Error(),
	}

	return &page
}

func (c *ErrorPage) Init() tea.Cmd {
	return nil
}

func (c *ErrorPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return c, PopPageCmd
		case "q":
			return c, ExitCmd
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	c.viewport, cmd = c.viewport.Update(msg)
	cmds = append(cmds, cmd)

	c.statusBar, cmd = c.statusBar.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *ErrorPage) SetSize(width, height int) {
	c.statusBar.Width = width
	c.viewport.Width = width

	availableHeight := max(0, height-lipgloss.Height(c.statusBar.View()))
	c.viewport.Height = availableHeight

	text := wordwrap.String(c.msg, width-2)
	c.viewport.SetContent(text)
}

func (c *ErrorPage) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.viewport.View(), c.statusBar.View())
}
