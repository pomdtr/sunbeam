package tui

import (
	"encoding/json"
	"fmt"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type RootList struct {
	w, h       int
	err        *Detail
	list       *List
	generator  func() ([]types.ListItem, error)
	OnSelect   func(id string)
	extensions map[string]Extension
}

func NewRootList(extensions map[string]Extension, generator func() ([]types.ListItem, error)) *RootList {
	return &RootList{
		list:       NewList(),
		generator:  generator,
		extensions: extensions,
	}
}

func (c *RootList) Init() tea.Cmd {
	return c.Reload()
}

func (c RootList) Reload() tea.Cmd {
	return tea.Sequence(c.list.SetIsLoading(true), func() tea.Msg {
		list, err := c.generator()
		if err != nil {
			return err
		}

		return list
	})
}

func (c *RootList) Focus() tea.Cmd {
	termOutput.SetWindowTitle("Sunbeam")
	return c.list.Focus()
}

func (c *RootList) Blur() tea.Cmd {
	return c.list.SetIsLoading(false)
}

func (c *RootList) SetSize(width, height int) {
	c.w, c.h = width, height
	if c.err != nil {
		c.err.SetSize(width, height)
	}
	c.list.SetSize(width, height)
}

func (c *RootList) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			item, ok := c.list.Selection()
			if !ok {
				break
			}

			if c.OnSelect != nil {
				c.OnSelect(item.Id)
			}
		case "ctrl+r":
			return c, c.Reload()
		case "ctrl+e":
			editCmd := utils.EditCmd(utils.ConfigPath())
			return c, tea.ExecProcess(editCmd, func(err error) tea.Msg {
				if err != nil {
					return err
				}

				return types.Command{
					Type: types.CommandTypeReload,
				}
			})
		}
	case []types.ListItem:
		c.list.SetItems(msg...)
		return c, c.list.SetIsLoading(false)
	case error:
		c.err = NewErrorPage(msg)
		c.err.SetSize(c.w, c.h)
		return c, c.err.Init()
	case types.Command:
		switch msg.Type {
		case types.CommandTypeRun:
			return c, func() tea.Msg {
				extension, ok := c.extensions[msg.Extension]
				if !ok {
					return fmt.Errorf("extension %s not found", msg.Extension)
				}

				command, ok := extension.Command(msg.Command)
				if !ok {
					return fmt.Errorf("command %s not found", msg.Command)
				}

				if command.Mode == types.CommandModeView {
					return PushPageMsg{NewRunner(c.extensions, types.CommandRef{
						Extension: msg.Extension,
						Command:   msg.Command,
						Params:    msg.Params,
					})}
				}

				out, err := extension.Run(types.CommandInput{
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
		case types.CommandTypeCopy:
			return c, func() tea.Msg {
				if err := clipboard.WriteAll(msg.Text); err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			}

		case types.CommandTypeOpen:
			return c, func() tea.Msg {
				if err := utils.OpenWith(msg.Target, msg.App); err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			}
		case types.CommandTypeReload:
			return c, c.Reload()
		case types.CommandTypeExit, types.CommandTypePop:
			return c, ExitCmd
		default:
			return c, nil
		}
	}

	if c.err != nil {
		page, cmd := c.err.Update(msg)
		c.err = page.(*Detail)
		return c, cmd
	}

	page, cmd := c.list.Update(msg)
	c.list = page.(*List)

	return c, cmd
}

func (c *RootList) View() string {
	if c.err != nil {
		return c.err.View()
	}
	return c.list.View()
}
