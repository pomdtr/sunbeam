package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
)

type Detail struct {
	header     Header
	Style      lipgloss.Style
	viewport   viewport.Model
	actions    ActionList
	content    string
	ready      bool
	ContentCmd func() string
	footer     Footer
}

func NewDetail(title string, contentCmd func() string) *Detail {
	footer := NewFooter(title)

	actionList := NewActionList()
	actionList.SetTitle(title)

	header := NewHeader()

	d := Detail{
		header:     header,
		actions:    actionList,
		ContentCmd: contentCmd,
		footer:     footer,
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
	return func() tea.Msg {
		content := d.ContentCmd()
		return PreviewContentMsg(content)
	}
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
			return &c, func() tea.Msg {
				return PopPageMsg{}
			}

		}
	case PreviewContentMsg:
		c.content = string(msg)
		c.RefreshPreview()
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

func (c *Detail) RefreshPreview() {
	content := wrap.String(wordwrap.String(c.content, c.viewport.Width-2), c.viewport.Width-2)
	c.viewport.SetContent(content)
}

func (c *Detail) SetSize(width, height int) {
	c.footer.Width = width
	c.header.Width = width
	c.actions.SetSize(width, height)

	viewportHeight := height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())
	if !c.ready {
		c.ready = true
		c.viewport = viewport.New(width, viewportHeight)
		c.viewport.Style = lipgloss.NewStyle().Padding(0, 1)
	} else {
		c.viewport.Width = width
		c.viewport.Height = viewportHeight
	}

	c.RefreshPreview()
}

func (c *Detail) View() string {
	if !c.ready {
		return ""
	}

	if c.actions.Focused() {
		return c.actions.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
