package tui

import (
	"encoding/json"
	"fmt"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/browser"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type RootList struct {
	list       *List
	extensions Extensions
	OnSelect   func(types.ListItem) error
}

func NewRootList(extensions Extensions, items ...types.ListItem) *RootList {
	page := RootList{
		extensions: extensions,
		list:       NewList(items...),
	}

	return &page
}

func (c *RootList) Init() tea.Cmd {
	return tea.Batch(c.list.Init(), FocusCmd)
}

func (c *RootList) SetSize(width, height int) {
	c.list.SetSize(width, height)
}

func (c *RootList) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if c.OnSelect == nil {
				break
			}

			item, ok := c.list.Selection()
			if !ok {
				break
			}

			if err := c.OnSelect(item); err != nil {
				return c, func() tea.Msg {
					return err
				}
			}
		}

	case types.Command:
		return c, func() tea.Msg {
			if msg.Type != types.CommandTypeRun {
				switch msg.Type {
				case types.CommandTypeCopy:
					if err := clipboard.WriteAll(msg.Text); err != nil {
						return err
					}

					if msg.Exit {
						return ExitMsg{}
					}

					return nil
				case types.CommandTypeOpen:
					if err := browser.OpenURL(msg.Text); err != nil {
						return err
					}

					if msg.Exit {
						return ExitMsg{}
					}

					return nil
				default:
					return nil
				}
			}

			extension, err := c.extensions.Get(msg.Origin)
			if err != nil {
				return err
			}

			command, ok := extension.Command(msg.Command)
			if !ok {
				return fmt.Errorf("command %s not found", msg.Command)
			}

			if command.Mode == types.CommandModeView {
				return PushPageMsg{NewRunner(c.extensions, CommandRef{
					Origin:  msg.Origin,
					Command: msg.Command,
					Params:  msg.Params,
				})}
			}

			out, err := extension.Run(CommandInput{
				Command: msg.Command,
				Params:  msg.Params,
			})
			if err != nil {
				return err
			}

			if len(out) == 0 {
				return ExitMsg{}
			}

			outputCommand := types.Command{}
			if err := json.Unmarshal(out, &outputCommand); err != nil {
				return err
			}

			return outputCommand
		}
	}

	page, cmd := c.list.Update(msg)
	c.list = page.(*List)

	return c, cmd
}

func (c *RootList) View() string {
	return c.list.View()
}
