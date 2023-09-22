package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type ExtensionPage struct {
	origin    string
	extension Extension
	list      *List
}

func NewExtensionPage(origin string) *ExtensionPage {
	list := NewList("Loading...")
	return &ExtensionPage{
		list:   list,
		origin: origin,
	}
}

func (c *ExtensionPage) Init() tea.Cmd {
	return tea.Batch(c.list.SetIsLoading(true), c.list.Init(), func() tea.Msg {
		extension, err := LoadExtension(c.origin)
		if err != nil {
			return err
		}
		return extension
	})
}

func (c *ExtensionPage) SetSize(width, height int) {
	c.list.SetSize(width, height)
}

func IsRootCommand(command types.Command) bool {
	for _, param := range command.Params {
		if !param.Optional {
			return false
		}
	}

	return true
}

func (c *ExtensionPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case Extension:
		extension := msg
		c.extension = extension

		var items []types.ListItem
		for _, command := range extension.Commands {
			if IsRootCommand(command) {
				items = append(items, types.ListItem{
					Title: command.Title,
					Actions: []types.Action{
						{
							Type:  types.ActionTypeRun,
							Title: "Run Command",
							Command: types.CommandRef{
								Origin: c.origin,
								Name:   command.Name,
								Params: make(map[string]any),
							},
						},
					},
				})
			}
		}

		c.list.footer.title = extension.Title
		c.list.SetItems(items...)
		return c, c.list.SetIsLoading(false)
	case types.Action:
		action := msg
		return c, func() tea.Msg {
			if action.Type == types.ActionTypeRun {
				return PushPageMsg{NewCommand(Extensions{c.origin: c.extension}, action.Command)}
			}

			if err := RunAction(action); err != nil {
				return err
			}

			if action.Exit {
				return ExitMsg{}
			}

			return nil
		}
	}

	var cmd tea.Cmd
	page, cmd := c.list.Update(msg)
	c.list = page.(*List)

	return c, cmd
}

func (c *ExtensionPage) View() string {
	return c.list.View()
}
