package tui

import (
	"bytes"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pomdtr/sunbeam/pkg/types"
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
	text       string
	language   string
}

func NewDetail(title string, text string, actions ...types.Action) *Detail {
	footer := NewFooter(title)
	if len(actions) == 1 {
		footer.SetBindings(
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", actions[0].Title)),
		)
	} else if len(actions) > 1 {
		footer.SetBindings(
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", actions[0].Title)),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "Actions")),
		)
	}

	actionList := NewActionList(actions...)
	viewport := viewport.New(0, 0)
	viewport.Style = lipgloss.NewStyle().Padding(0, 1)
	header := NewHeader()

	d := Detail{
		viewport:   viewport,
		header:     header,
		actionList: actionList,
		footer:     footer,
		text:       text,
	}

	d.RefreshContent()
	return &d
}

func (d *Detail) Init() tea.Cmd {
	return nil
}

type DetailMsg string

func (d *Detail) SetIsLoading(isLoading bool) tea.Cmd {
	return d.header.SetIsLoading(isLoading)
}

func (c *Detail) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
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
				break
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

			return c, func() tea.Msg {
				actions := c.actionList.actions
				if len(actions) == 0 {
					return nil
				}

				return actions[0]
			}
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
	writer := bytes.Buffer{}
	text := c.text
	if c.language != "" {
		err := quick.Highlight(&writer, c.text, c.language, "terminal16", theme)
		if err != nil {
			return err
		}
		text = writer.String()
	}

	c.viewport.SetContent(wordwrap.String(text, c.viewport.Width-2))
	return nil
}

func (c *Detail) SetSize(width, height int) {
	c.footer.Width = width
	c.header.Width = width
	c.viewport.Width = width
	c.actionList.SetSize(width, height)

	c.viewport.Height = height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())

	c.RefreshContent()
}

func (c *Detail) View() string {
	if c.actionList.Focused() {
		return c.actionList.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
