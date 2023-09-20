package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type RootList struct {
	header Header
	footer Footer
	filter Filter

	extensions Extensions
}

func NewRootPage(extensions Extensions, allowlist ...string) *RootList {
	items := make([]FilterItem, 0)

	allowMap := make(map[string]bool)
	for _, alias := range allowlist {
		allowMap[alias] = true
	}

	for alias, ext := range extensions {
		for i, item := range ext.Items {
			if _, ok := allowMap[alias]; len(allowMap) > 0 && !ok {
				continue
			}

			command, ok := ext.Commands[item.Command]
			if !ok {
				continue
			}

			title := item.Title
			if title == "" {
				title = command.Title
			}

			items = append(items, ListItem{
				Id:          fmt.Sprintf("%s:%d", alias, i),
				Title:       title,
				Subtitle:    ext.Title,
				Accessories: []string{alias},
				Actions: []types.Action{
					{
						Type: types.ActionTypeRun,
						Text: "Run",
						Command: types.CommandRef{
							Extension: alias,
							Name:      item.Command,
							Params:    item.Params,
						},
					},
				},
			})

		}
	}

	filter := NewFilter(items...)
	filter.DrawLines = true
	footer := NewFooter("Sunbeam")
	footer.SetBindings(
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", "Run")),
	)

	page := RootList{
		extensions: extensions,
		header:     NewHeader(),
		footer:     footer,
		filter:     filter,
	}

	return &page
}

func (c *RootList) Init() tea.Cmd {
	return tea.Batch(c.header.Init(), FocusCmd)
}

func (c *RootList) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return c, PopPageCmd
		case "enter":
			selection := c.filter.Selection()
			if selection == nil {
				return c, nil
			}

			item, ok := selection.(ListItem)
			if !ok {
				return c, nil
			}

			if len(item.Actions) == 0 {
				return c, nil
			}

			action := item.Actions[0]

			return c, PushPageCmd(NewCommand(c.extensions, action.Command))
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	header, cmd := c.header.Update(msg)
	cmds = append(cmds, cmd)

	filter, cmd := c.filter.Update(msg)
	cmds = append(cmds, cmd)

	if header.Value() != c.header.Value() {
		filter.FilterItems(header.Value())
	}

	c.header = header
	c.filter = filter
	return c, tea.Batch(cmds...)
}

func (c *RootList) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.filter.View(), c.footer.View())
}

func (c *RootList) SetSize(width, height int) {
	c.header.Width = width
	c.footer.Width = width
	c.filter.Width = width

	c.filter.Height = height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())
}
