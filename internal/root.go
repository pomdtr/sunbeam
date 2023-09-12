package internal

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg"
)

type RootList struct {
	header Header
	footer Footer
	filter Filter

	extensions map[string]Extension
}

func NewRootPage(extensions map[string]Extension) *RootList {
	items := make([]FilterItem, 0)

	for alias, ext := range extensions {
		for _, command := range ext.Commands {
			if len(command.Params) > 0 {
				continue
			}
			var primaryAction pkg.Action
			if command.Mode == "silent" {
				primaryAction = pkg.Action{
					Type: pkg.ActionTypeRun,
					Text: "Run",
					Command: pkg.CommandRef{
						Alias: alias,
						Name:  command.Name,
					},
				}
			} else {
				primaryAction = pkg.Action{
					Type: pkg.ActionTypePush,
					Text: "Push",
					Command: pkg.CommandRef{
						Alias: alias,
						Name:  command.Name,
					},
				}
			}
			items = append(items, ListItem{
				Id:          fmt.Sprintf("%s:%s", alias, command.Name),
				Title:       command.Title,
				Subtitle:    ext.Title,
				Accessories: []string{alias},
				Actions: []pkg.Action{
					primaryAction,
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
			alias := action.Command.Alias
			extension, ok := c.extensions[alias]
			if !ok {
				return c, nil
			}

			return c, runAction(action, extension)
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
