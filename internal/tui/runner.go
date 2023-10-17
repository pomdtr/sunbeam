package tui

import (
	"encoding/json"
	"fmt"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Runner struct {
	embed         Page
	width, height int

	extension extensions.Extension
	command   types.CommandSpec
	input     types.CommandInput
}

func NewRunner(extension extensions.Extension, command types.CommandSpec, input types.CommandInput) *Runner {
	return &Runner{
		embed:     NewDetail(""),
		extension: extension,
		command:   command,
		input:     input,
	}
}

func (c *Runner) SetIsLoading(isLoading bool) tea.Cmd {
	switch page := c.embed.(type) {
	case *Detail:
		return page.SetIsLoading(isLoading)
	case *List:
		return page.SetIsLoading(isLoading)
	}

	return nil
}

func (c *Runner) Init() tea.Cmd {
	return tea.Batch(c.Reload(), c.embed.Init())
}

func (c *Runner) Focus() tea.Cmd {
	if c.embed == nil {
		return nil
	}
	termenv.DefaultOutput().SetWindowTitle(fmt.Sprintf("%s - %s", c.command.Title, c.extension.Title))
	return c.embed.Focus()
}

func (c *Runner) Blur() tea.Cmd {
	return nil
}

func (c *Runner) SetSize(w int, h int) {
	c.width = w
	c.height = h

	c.embed.SetSize(w, h)
}

func (c *Runner) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if c.embed != nil {
				break
			}
			return c, PopPageCmd
		case "ctrl+r":
			return c, c.Reload()
		}

	case Page:
		c.embed = msg
		c.embed.SetSize(c.width, c.height)
		return c, tea.Sequence(c.embed.Init(), c.embed.Focus())
	case types.Command:
		switch msg.Type {
		case types.CommandTypeRun:
			command, ok := c.extension.Command(msg.Command)
			if !ok {
				c.embed = NewErrorPage(fmt.Errorf("command %s not found", msg.Command))
				c.embed.SetSize(c.width, c.height)
				return c, c.embed.Init()
			}

			switch command.Mode {
			case types.CommandModeView:
				runner := NewRunner(c.extension, command, types.CommandInput{
					Params: msg.Params,
				})

				return c, PushPageCmd(runner)
			case types.CommandModeNoView:
				return c, func() tea.Msg {
					output, err := c.extension.Run(command.Name, types.CommandInput{
						Params: msg.Params,
					})
					if err != nil {
						return err
					}

					if err := schemas.ValidateCommand(output); err != nil {
						return err
					}

					var res types.Command
					if err := json.Unmarshal(output, &res); err != nil {
						return err
					}

					return res
				}

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
			command := msg
			return c, func() tea.Msg {
				if err := utils.OpenWith(command.Target, command.App); err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			}
		case types.CommandTypeReload:
			if c.input.Params == nil {
				c.input.Params = make(map[string]any)
			}

			for k, v := range msg.Params {
				c.input.Params[k] = v
			}

			return c, c.Reload()
		case types.CommandTypeExit:
			return c, ExitCmd
		}
	case error:
		c.embed = NewErrorPage(msg)
		c.embed.SetSize(c.width, c.height)
		return c, c.embed.Init()
	}

	var cmd tea.Cmd
	c.embed, cmd = c.embed.Update(msg)
	return c, cmd
}

func (c *Runner) View() string {
	return c.embed.View()
}

func (c *Runner) Reload() tea.Cmd {
	return tea.Sequence(c.SetIsLoading(true), func() tea.Msg {
		output, err := c.extension.Run(c.command.Name, c.input)
		if err != nil {
			return err
		}

		if err := schemas.ValidateView(output); err != nil {
			return err
		}

		var view types.View
		if err := json.Unmarshal(output, &view); err != nil {
			return err
		}

		if view.Title != "" {
			termenv.DefaultOutput().SetWindowTitle(fmt.Sprintf("%s - %s", view.Title, c.extension.Title))
		} else {
			termenv.DefaultOutput().SetWindowTitle(fmt.Sprintf("%s - %s", c.command.Title, c.extension.Title))
		}

		switch view.Type {
		case types.ViewTypeDetail:
			var detail types.Detail
			if err := json.Unmarshal(output, &detail); err != nil {
				return err
			}

			return NewDetail(detail.Markdown, detail.Actions...)
		case types.ViewTypeList:
			var list types.List
			if err := json.Unmarshal(output, &list); err != nil {
				return err
			}

			page := NewList(list.Items...)
			if list.EmptyText != "" {
				page.SetEmptyText(list.EmptyText)
			}
			if list.Actions != nil {
				page.SetActions(list.Actions...)
			}

			if list.Dynamic {
				page.OnQueryChange = func(query string) tea.Cmd {
					c.input.Query = query
					return c.Reload()
				}
			}

			page.SetSize(c.width, c.height)
			if embed, ok := c.embed.(*List); ok {
				page.SetQuery(embed.Query())
			}

			return page
		default:
			return fmt.Errorf("invalid view type")
		}
	})
}
