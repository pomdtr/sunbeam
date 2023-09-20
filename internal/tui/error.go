package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type ErrorPage struct {
	header   Header
	footer   Footer
	viewport viewport.Model
	msg      string
}

func NewErrorPage(err error) *ErrorPage {
	viewport := viewport.New(0, 0)
	viewport.SetContent(err.Error())
	page := ErrorPage{
		header:   NewHeader(),
		footer:   NewFooter("Error"),
		viewport: viewport,
		msg:      err.Error(),
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

	c.header, cmd = c.header.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *ErrorPage) SetSize(width, height int) {
	c.header.Width = width
	c.footer.Width = width
	c.viewport.Width = width

	availableHeight := max(0, height-lipgloss.Height(c.header.View())-lipgloss.Height(c.footer.View()))
	c.viewport.Height = availableHeight

	text := wordwrap.String(c.msg, width-2)
	c.viewport.SetContent(text)
}

func (c *ErrorPage) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
