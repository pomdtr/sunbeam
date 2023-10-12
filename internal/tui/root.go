package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type RootList struct {
	w, h       int
	title      string
	err        *Detail
	list       *List
	generator  func() (map[string]extensions.Extension, []types.ListItem, error)
	OnSelect   func(id string)
	extensions map[string]extensions.Extension
}

func NewRootList(title string, generator func() (map[string]extensions.Extension, []types.ListItem, error)) *RootList {
	return &RootList{
		title:     title,
		list:      NewList(),
		generator: generator,
	}
}

func (c *RootList) Init() tea.Cmd {
	return c.Reload()
}

type RefreshMsg struct {
	extensions map[string]extensions.Extension
	rootItems  []types.ListItem
}

func (c RootList) Reload() tea.Cmd {
	return tea.Sequence(c.list.SetIsLoading(true), func() tea.Msg {
		extensions, items, err := c.generator()
		if err != nil {
			return err
		}

		return RefreshMsg{
			extensions: extensions,
			rootItems:  items,
		}
	})
}

func (c *RootList) Focus() tea.Cmd {
	termenv.NewOutput(os.Stdout).SetWindowTitle(c.title)
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

func (c *RootList) SetError(err error) tea.Cmd {
	c.err = NewErrorPage(err)
	c.err.SetSize(c.w, c.h)
	return func() tea.Msg {
		return err
	}
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
		}
	case RefreshMsg:
		c.list.SetItems(msg.rootItems...)
		c.extensions = msg.extensions
		return c, c.list.SetIsLoading(false)
	case error:
		c.err = NewErrorPage(msg)
		c.err.SetSize(c.w, c.h)
		return c, c.err.Init()
	case types.Command:
		switch msg.Type {
		case types.CommandTypeRun:
			extension, ok := c.extensions[msg.Extension]
			if !ok {
				return c, c.SetError(fmt.Errorf("extension %s not found", msg.Extension))
			}

			commandspec, ok := extension.Command(msg.Command)
			if !ok {
				return c, c.SetError(fmt.Errorf("command %s not found", msg.Command))
			}

			switch commandspec.Mode {
			case types.CommandModeView:
				runner, err := NewRunner(c.extensions, CommandRef{
					Extension: msg.Extension,
					Command:   msg.Command,
					Params:    msg.Params,
				})

				if err != nil {
					return c, c.SetError(err)
				}
				return c, PushPageCmd(runner)
			case types.CommandModeNoView:
				return c, func() tea.Msg {
					out, err := extension.Run(commandspec.Name, types.CommandInput{
						Params: msg.Params,
					})
					if err != nil {
						return err
					}

					if len(out) == 0 {
						return ExitCmd()
					}

					var command types.Command
					if err := json.Unmarshal(out, &command); err != nil {
						return err
					}
					return command
				}
			case types.CommandModeTTY:
				cmd, err := extension.Cmd(commandspec.Name, types.CommandInput{
					Params: msg.Params,
				})
				if err != nil {
					return c, c.SetError(err)
				}

				output := bytes.Buffer{}
				cmd.Stdout = &output

				return c, tea.ExecProcess(cmd, func(err error) tea.Msg {
					if err != nil {
						return err
					}

					var command types.Command
					if err := json.Unmarshal(output.Bytes(), &command); err != nil {
						return err
					}

					return command
				})
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
