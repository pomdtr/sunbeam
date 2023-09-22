package tui

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/alessio/shellescape"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type RootList struct {
	list *List
}

func shellCommand(item types.RootItem) (string, error) {
	args := []string{"sunbeam"}
	if item.Extension != "" {
		args = append(args, item.Extension)
	} else {
		args = append(args, "run", item.Origin)
	}

	args = append(args, item.Command)
	for key, value := range item.Params {
		switch v := value.(type) {
		case string:
			args = append(args, fmt.Sprintf("--%s=%s", key, v))
		case bool:
			if v {
				args = append(args, fmt.Sprintf("--%s", key))
			}
		default:
			return "", fmt.Errorf("unknown type %T", v)
		}
	}

	return shellescape.QuoteCommand(args), nil
}

func NewRootList(title string, rootItems ...types.RootItem) *RootList {
	items := make([]types.ListItem, 0)
	for _, rootitem := range rootItems {
		shell, err := shellCommand(rootitem)
		if err != nil {
			continue
		}

		listitem := types.ListItem{
			Id:    shell,
			Title: rootitem.Title,
			Actions: []types.Action{
				{
					Type:  types.ActionTypeRun,
					Title: "Run Command",
					Command: types.CommandRef{
						Origin: rootitem.Origin,
						Name:   rootitem.Command,
						Params: rootitem.Params,
					},
				},
			},
		}

		if rootitem.Extension != "" {
			var accessory string
			if rootitem.Command != "" {
				accessory = fmt.Sprintf("%s %s", rootitem.Extension, rootitem.Command)
			} else {
				accessory = rootitem.Extension
			}
			listitem.Accessories = []string{accessory}
			listitem.Actions = append(listitem.Actions, types.Action{
				Type:  types.ActionTypeCopy,
				Title: "Copy as Shell Command",
				Text:  shell,
				Exit:  true,
			})

			buffer := bytes.Buffer{}
			encoder := json.NewEncoder(&buffer)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(rootitem); err != nil {
				continue
			}

			listitem.Actions = append(listitem.Actions, types.Action{
				Type:  types.ActionTypeCopy,
				Title: "Copy as JSON",
				Text:  buffer.String(),
				Exit:  true,
			})
		}

		items = append(items, listitem)
	}
	list := NewList(title, items...)
	page := RootList{
		list: list,
	}

	return &page
}

func (c *RootList) Init() tea.Cmd {
	return tea.Batch(c.list.Init(), FocusCmd)
}

func (c *RootList) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case types.Action:
		action := msg
		return c, func() tea.Msg {
			if action.Type == types.ActionTypeRun {
				if action.Command.Name == "" {
					return PushPageMsg{NewExtensionPage(action.Command.Origin)}
				}

				return PushPageMsg{NewCommand(make(Extensions), action.Command)}
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

	page, cmd := c.list.Update(msg)
	c.list = page.(*List)

	return c, cmd
}
func (c *RootList) View() string {
	return c.list.View()
}

func (c *RootList) SetSize(width, height int) {
	c.list.SetSize(width, height)
}
