package tui

import (
	"bytes"
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

type ReloadMsg struct {
	params map[string]any
}

func ReloadCmd(params map[string]any) tea.Cmd {
	return func() tea.Msg {
		return ReloadMsg{
			params: params,
		}
	}
}

func NewRunner(extension extensions.Extension, command types.CommandSpec, input types.CommandInput) *Runner {
	return &Runner{
		extension: extension,
		embed:     NewDetail(""),
		command:   command,
		input:     input,
	}
}

func (c *Runner) SetIsLoading(isLoading bool) tea.Cmd {
	if c.embed == nil {
		return nil
	}

	switch page := c.embed.(type) {
	case *Detail:
		return page.SetIsLoading(isLoading)
	case *List:
		return page.SetIsLoading(isLoading)
	case *Form:
		return page.SetIsLoading(isLoading)
	}

	return nil
}

func (c *Runner) Init() tea.Cmd {
	return tea.Batch(c.Reload(c.input), c.embed.Init())
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

	if c.embed != nil {
		c.embed.SetSize(w, h)
	}
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
			return c, c.Reload(c.input)
		}
	case types.Form:
		form := msg
		page, err := NewForm(c.extension.Title, form.Fields...)
		if err != nil {
			c.embed = NewErrorPage(err)
			c.embed.SetSize(c.width, c.height)
			return c, c.embed.Init()
		}

		page.SetSize(c.width, c.height)
		c.embed = page
		return c, tea.Sequence(c.embed.Init(), c.embed.Focus())
	case types.Detail:
		detail := msg

		page := NewDetail(detail.Markdown, detail.Actions...)
		c.embed = page
		c.embed.SetSize(c.width, c.height)
		return c, tea.Sequence(c.embed.Init(), c.embed.Focus())
	case types.List:
		list := msg
		page := NewList(list.Items...)
		if list.EmptyText != "" {
			page.SetEmptyText(list.EmptyText)
		}
		if list.Actions != nil {
			page.SetActions(list.Actions...)
		}

		if list.Reload {
			page.OnQueryChange = func(query string) tea.Cmd {
				c.input.Query = query
				return c.Reload(c.input)
			}
		}

		page.SetSize(c.width, c.height)
		if c.embed == nil {
			c.embed = page
			return c, tea.Sequence(c.embed.Init(), c.embed.Focus())
		}

		if embed, ok := c.embed.(*List); ok {
			page.SetQuery(embed.Query())
		}

		c.embed = page
		return c, tea.Sequence(c.embed.Init(), c.embed.Focus())
	case SubmitMsg:
		return c, func() tea.Msg {
			input := c.input
			input.FormData = msg
			output, err := c.extension.Run(c.command.Name, input)
			if err != nil {
				return err
			}

			if len(output) == 0 {
				return ExitMsg{}
			}

			var command types.Command
			if err := json.Unmarshal(output, &command); err != nil {
				return err
			}

			return command
		}
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
			case types.CommandModeNoView:
				return c, func() tea.Msg {
					output, err := c.extension.Run(
						command.Name,
						types.CommandInput{Params: msg.Params},
					)
					if err != nil {
						return err
					}

					if len(output) == 0 {
						return nil
					}

					if err := schemas.ValidateCommand(output); err != nil {
						return err
					}

					var command types.Command
					if err := json.Unmarshal(output, &command); err != nil {
						return err
					}

					return command
				}
			case types.CommandModeView:
				command, ok := c.extension.Command(msg.Command)
				if !ok {
					return c, func() tea.Msg {
						return fmt.Errorf("command %s not found", msg.Command)
					}
				}

				return c, PushPageCmd(NewRunner(c.extension, command, types.CommandInput{
					Params: msg.Params,
				}))
			case types.CommandModeTTY:
				cmd, err := c.extension.Cmd(command.Name, types.CommandInput{Params: msg.Params})
				if err != nil {
					return c, func() tea.Msg {
						return err
					}
				}

				output := bytes.Buffer{}
				cmd.Stdout = &output
				return c, tea.ExecProcess(cmd, func(err error) tea.Msg {
					if err != nil {
						return err
					}

					if len(output.Bytes()) == 0 {
						return nil
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
			return c, c.Reload(types.CommandInput{
				Params: msg.Params,
			})
		case types.CommandTypeExit:
			return c, ExitCmd
		case types.CommandTypePop:
			if msg.Reload {
				return c, tea.Sequence(PopPageCmd, ReloadCmd(nil))
			}
			return c, PopPageCmd
		}
	case ReloadMsg:
		return c, c.Reload(types.CommandInput{
			Params: msg.params,
		})
	case error:
		c.embed = NewErrorPage(msg)
		c.embed.SetSize(c.width, c.height)
		return c, c.embed.Init()
	}

	if c.embed != nil {
		var cmd tea.Cmd
		c.embed, cmd = c.embed.Update(msg)
		return c, cmd
	}

	return c, nil
}

func (c *Runner) View() string {
	return c.embed.View()
}

func (c *Runner) Reload(input types.CommandInput) tea.Cmd {
	return tea.Sequence(c.SetIsLoading(true), func() tea.Msg {
		output, err := c.extension.Run(c.command.Name, input)
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

			return detail
		case types.ViewTypeList:
			var list types.List
			if err := json.Unmarshal(output, &list); err != nil {
				return err
			}

			return list
		case types.ViewTypeForm:
			var form types.Form
			if err := json.Unmarshal(output, &form); err != nil {
				return err
			}

			return form
		default:
			return fmt.Errorf("invalid command output")
		}
	})
}
