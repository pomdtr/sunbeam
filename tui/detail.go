package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type Detail struct {
	header     Header
	Style      lipgloss.Style
	content    string
	viewport   viewport.Model
	actionList ActionList
	footer     Footer
}

func NewDetail(title string) *Detail {
	viewport := viewport.New(0, 0)
	viewport.Style = styles.Regular.Copy()

	footer := NewFooter(title)

	actionList := NewActionList()
	actionList.SetTitle(title)

	header := NewHeader()

	d := Detail{
		viewport:   viewport,
		header:     header,
		actionList: actionList,
		footer:     footer,
	}

	return &d
}

func (c *Detail) SetActions(actions ...Action) {
	c.actionList.SetActions(actions...)

	if len(actions) > 0 {
		c.footer.SetBindings(
			actions[0].Binding(),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Actions")),
		)
	}
}

func (d *Detail) Init() tea.Cmd {
	return nil
}

func (d *Detail) SetContent(content string) {
	d.content = content
	content = wordwrap.String(content, d.viewport.Width-4)
	content = lipgloss.NewStyle().Padding(0, 2).Width(d.viewport.Width).Render(content)
	d.viewport.SetContent(content)
}

func (d *Detail) SetIsLoading(isLoading bool) tea.Cmd {
	return d.header.SetIsLoading(isLoading)
}

func (c Detail) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			switch msg.String() {
			case "q", "Q":
				return &c, tea.Quit
			}
		case tea.KeyEscape:
			if c.actionList.Focused() {
				break
			}
			return &c, PopCmd
		}
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd

	c.viewport, cmd = c.viewport.Update(msg)
	cmds = append(cmds, cmd)

	c.actionList, cmd = c.actionList.Update(msg)
	cmds = append(cmds, cmd)

	c.header, cmd = c.header.Update(msg)
	cmds = append(cmds, cmd)

	return &c, tea.Batch(cmds...)
}

func (c *Detail) SetSize(width, height int) {
	c.footer.Width = width
	c.header.Width = width
	c.viewport.Width = width
	c.viewport.Height = height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())
	c.actionList.SetSize(width, height)

	c.SetContent(c.content)
}

func (c *Detail) View() string {
	if c.actionList.Focused() {
		return c.actionList.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
