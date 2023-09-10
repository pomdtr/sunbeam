package internal

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/pomdtr/sunbeam/pkg"
)

type Detail struct {
	header     Header
	Style      lipgloss.Style
	viewport   viewport.Model
	actionList ActionList
	ready      bool
	footer     Footer
	raw        string

	generator func() (pkg.Detail, error)
	runner    func(pkg.Action) tea.Cmd
}

func NewDetail(title string, generator func() (pkg.Detail, error), runner func(pkg.Action) tea.Cmd) *Detail {
	footer := NewFooter(title)

	actionList := NewActionList(runner)
	actionList.SetTitle(title)
	viewport := viewport.New(0, 0)

	header := NewHeader()

	d := Detail{
		viewport:   viewport,
		header:     header,
		actionList: actionList,
		footer:     footer,

		generator: generator,
		runner:    runner,
	}

	return &d
}
func (d *Detail) Init() tea.Cmd {
	return tea.Batch(d.header.SetIsLoading(true), d.Reload)
}

func (d *Detail) Focus() tea.Cmd {
	return nil
}

func (d *Detail) Reload() tea.Msg {
	detail, err := d.generator()
	if err != nil {
		return err
	}

	return detail
}

type DetailMsg string

func (d *Detail) SetIsLoading(isLoading bool) tea.Cmd {
	return d.header.SetIsLoading(isLoading)
}

func (c *Detail) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case pkg.Detail:
		c.SetIsLoading(false)
		c.RefreshContent(msg.Text)
		c.actionList.SetActions(msg.Actions...)
		if len(msg.Actions) == 1 {
			c.footer.SetBindings(
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", msg.Actions[0].Title)),
			)
		} else if len(msg.Actions) > 1 {
			c.footer.SetBindings(
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", msg.Actions[0].Title)),
				key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "Actions")),
			)
		}

		if msg.Title != "" {
			c.footer.title = msg.Title
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if c.actionList.Focused() {
				break
			}

			if len(c.actionList.actions) == 0 {
				return c, nil
			}
			return c, c.actionList.Focus()

		case "esc":
			if c.actionList.Focused() {
				break
			}
			return c, func() tea.Msg {
				return PopPageMsg{}
			}
		case "enter":
			if c.actionList.Focused() {
				break
			}

			actions := c.actionList.actions
			if len(actions) == 0 {
				return c, nil
			}

			return c, c.runner(actions[0])
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

	return c, tea.Batch(cmds...)
}

func (c *Detail) RefreshContent(text string) {
	c.raw = text
	content := wrap.String(wordwrap.String(c.raw, c.viewport.Width-2), c.viewport.Width-2)
	c.viewport.SetContent(content)
}

func (c *Detail) SetSize(width, height int) {
	c.footer.Width = width
	c.header.Width = width
	c.actionList.SetSize(width, height)

	viewportHeight := height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())
	if !c.ready {
		c.ready = true
		c.viewport = viewport.New(width, viewportHeight)
		c.viewport.Style = lipgloss.NewStyle().Padding(0, 1)
	} else {
		c.viewport.Width = width
		c.viewport.Height = viewportHeight
	}

	c.RefreshContent(c.raw)
}

func (c *Detail) View() string {
	if !c.ready {
		return ""
	}

	if c.actionList.Focused() {
		return c.actionList.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
