package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var markdownRenderer *glamour.TermRenderer

func init() {
	var err error
	markdownRenderer, err = glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		log.Fatalf("failed to create markdown renderer: %v", err)
	}
}

type Detail struct {
	width, height int

	viewport.Model
	format  string
	actions []Action
	footer  *Footer
}

func NewDetail(format string, actions []Action) *Detail {
	viewport := viewport.New(0, 0)
	footer := NewFooter(actions...)
	return &Detail{Model: viewport, format: format, footer: footer}
}

func (d *Detail) SetContent(content string) error {
	switch d.format {
	case "markdown":
		rendered, err := markdownRenderer.Render(content)
		if err != nil {
			return err
		}
		d.Model.SetContent(rendered)
	default:
		d.Model.SetContent(content)
	}
	return nil
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
				if action.Shortcut == msg.String() {
					return c, action.Exec()
				}
			}

		}

	}
	var cmd tea.Cmd
	c.Model, cmd = c.Model.Update(msg)
	return c, cmd
}

func (c *Detail) SetSize(width, height int) {
	c.height = height
	c.width = width
	c.footer.Width = width
	c.Model.Width = width
	c.Model.Height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footer.View())
}

func (c *Detail) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.Model.View(), c.footer.View())
}

func (c *Detail) headerView() string {
	return SunbeamHeader(c.width)
}
