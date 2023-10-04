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
	extensions map[string]Extension
	OnSelect   func(id string)
}

func NewRootList(extensions map[string]Extension, items ...types.ListItem) *RootList {
	page := RootList{
		extensions: extensions,
		list:       NewList(items...),
	}

	return &page
}

func (c *RootList) Init() tea.Cmd {
	return tea.Batch(c.list.Init())
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
		}
	case error:
		c.err = NewErrorPage(msg)
		c.err.SetSize(c.w, c.h)
		return c, c.err.Init()
	case types.Command:
		return c, tea.Sequence(c.list.SetIsLoading(true), func() tea.Msg {
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
					command := msg
					if err := utils.OpenWith(command.Target, command.App); err != nil {
						return err
					}

					if msg.Exit {
						return ExitMsg{}
					}

					return nil
				case types.CommandTypeExit:
					return ExitMsg{}
				default:
					return nil
				}
			}

			extension, ok := c.extensions[msg.Extension]
			if !ok {
				return fmt.Errorf("extension %s not found", msg.Extension)
			}

			command, ok := extension.Command(msg.Command)
			if !ok {
				return fmt.Errorf("command %s not found", msg.Command)
			}

			if command.Mode == types.CommandModeView {
				return PushPageMsg{NewRunner(extension, command, msg.Params)}
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
		})
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
