package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Detail struct {
	header         Header
	Style          lipgloss.Style
	viewport       viewport.Model
	actions        ActionList
	PreviewCommand func() string
	footer         Footer
}

func NewDetail(title string) *Detail {
	viewport := viewport.New(0, 0)
	viewport.Style = lipgloss.NewStyle().Padding(0, 1)

	footer := NewFooter(title)

	actionList := NewActionList()
	actionList.SetTitle(title)

	header := NewHeader()

	d := Detail{
		viewport: viewport,
		header:   header,
		actions:  actionList,
		footer:   footer,
	}

	return &d
}

func (c *Detail) SetActions(actions ...Action) {
	c.actions.SetActions(actions...)

	if len(actions) == 0 {
		c.footer.SetBindings()
	} else {
		c.footer.SetBindings(
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", actions[0].Title)),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Show Actions")),
		)
	}
}

func (d *Detail) Init() tea.Cmd {
	if d.PreviewCommand != nil {
		return tea.Sequence(d.SetIsLoading(true), func() tea.Msg {
			content := d.PreviewCommand()
			return PreviewContentMsg(content)
		})
	}

	return nil
}

type DetailMsg string

func (d *Detail) SetIsLoading(isLoading bool) tea.Cmd {
	return d.header.SetIsLoading(isLoading)
}

func (c Detail) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			switch msg.String() {
			case "q", "Q":
				return &c, tea.Quit
			}
		case tea.KeyEscape:
			if c.actions.Focused() {
				break
			}
			return &c, PopCmd

		}
	case PreviewContentMsg:
		c.SetIsLoading(false)
		c.viewport.SetContent(string(msg))
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd

	c.viewport, cmd = c.viewport.Update(msg)
	cmds = append(cmds, cmd)

	c.actions, cmd = c.actions.Update(msg)
	cmds = append(cmds, cmd)

	c.header, cmd = c.header.Update(msg)
	cmds = append(cmds, cmd)

	return &c, tea.Batch(cmds...)
}

func (c *Detail) SetSize(width, height int) {
	c.footer.Width = width
	c.header.Width = width
	c.actions.SetSize(width, height)

	viewportHeight := height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())

	c.viewport.Width = width
	c.viewport.Height = viewportHeight
}

func (c *Detail) View() string {
	if c.actions.Focused() {
		return c.actions.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
