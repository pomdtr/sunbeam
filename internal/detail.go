package internal

import (
	"bytes"
	"encoding/json"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pomdtr/sunbeam/pkg"
)

var theme string

func init() {
	if lipgloss.HasDarkBackground() {
		theme = "monokai"
	} else {
		theme = "monokailight"
	}
}

type Detail struct {
	header     Header
	Style      lipgloss.Style
	viewport   viewport.Model
	actionList ActionList
	footer     Footer

	page      *pkg.Detail
	extension Extension
	command   pkg.Command
	params    pkg.CommandParams
}

func NewDetail(extension Extension, command pkg.Command, params pkg.CommandParams) *Detail {
	footer := NewFooter(command.Title)

	actionList := NewActionList()
	viewport := viewport.New(0, 0)
	viewport.Style = lipgloss.NewStyle().Padding(0, 1)

	header := NewHeader()

	d := Detail{
		viewport:   viewport,
		header:     header,
		actionList: actionList,
		footer:     footer,

		extension: extension,
		command:   command,
		params:    params,
	}

	return &d
}
func (d *Detail) Init() tea.Cmd {
	return tea.Batch(d.header.SetIsLoading(true), d.Reload)
}

func (d *Detail) Reload() tea.Msg {
	output, err := d.extension.Run(d.command.Name, pkg.CommandInput{
		Params: d.params,
	})
	if err != nil {
		return err
	}

	if err := pkg.ValidatePage(output); err != nil {
		return err
	}

	var page pkg.Detail
	if err := json.Unmarshal(output, &page); err != nil {
		return err
	}

	return page
}

type DetailMsg string

func (d *Detail) SetIsLoading(isLoading bool) tea.Cmd {
	return d.header.SetIsLoading(isLoading)
}

func (c *Detail) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case pkg.Detail:
		c.page = &msg
		c.SetIsLoading(false)
		c.RefreshContent()
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
		case "q":
			if c.actionList.Focused() {
				break
			}

			return c, func() tea.Msg {
				return ExitMsg{}
			}
		case "tab":
			if c.actionList.Focused() {
				break
			}

			if len(c.actionList.actions) == 0 {
				return c, nil
			}
			return c, nil

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

			return c, runAction(actions[0], c.extension)
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

func (c *Detail) RefreshContent() error {
	if c.page == nil {
		return nil
	}

	writer := bytes.Buffer{}
	err := quick.Highlight(&writer, c.page.Text, c.page.Language, "terminal16", theme)
	if err != nil {
		return err
	}

	text := wordwrap.String(writer.String(), c.viewport.Width-2)

	c.viewport.SetContent(text)

	return nil
}

func (c *Detail) SetSize(width, height int) {
	c.footer.Width = width
	c.header.Width = width
	c.viewport.Width = width
	c.actionList.SetSize(width, height)

	c.viewport.Height = height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())
}

func (c *Detail) View() string {
	if c.actionList.Focused() {
		return c.actionList.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
