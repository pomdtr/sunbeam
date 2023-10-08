package tui

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Runner struct {
	embed         Page
	width, height int

	alias      string
	extensions map[string]Extension
	extension  Extension
	command    types.CommandSpec
	params     map[string]any
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

func NewRunner(extensions map[string]Extension, ref types.CommandRef) (*Runner, error) {
	extension, ok := extensions[ref.Extension]
	if !ok {
		return nil, fmt.Errorf("extension %s not found", ref.Extension)
	}

	command, ok := extension.Command(ref.Command)
	if !ok {
		return nil, fmt.Errorf("command %s not found", ref.Command)
	}

	return &Runner{
		extensions: extensions,
		extension:  extensions[ref.Extension],
		embed:      NewDetail(""),
		alias:      ref.Extension,
		command:    command,
		params:     ref.Params,
	}, nil
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
	return tea.Batch(c.Reload(types.CommandInput{
		Params: c.params,
	}), c.embed.Init())
}

func (c *Runner) Focus() tea.Cmd {
	if c.embed == nil {
		return nil
	}
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
			return c, c.Reload(
				types.CommandInput{
					Params: c.params,
				},
			)
		}
	case types.Form:
		form := msg
		var formitems []FormItem
		for _, item := range form.Items {
			formitems = append(formitems, *NewFormItem(item))
		}

		page := NewForm(c.extension.Title, formitems...)
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
		if list.Reload {
			page.OnQueryChange = func(query string) tea.Cmd {
				return c.Reload(types.CommandInput{
					Query: query,
					Params: map[string]any{
						"query": query,
					},
				})
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
			output, err := c.extension.Run(c.command.Name, types.CommandInput{
				Params:   c.params,
				FormData: msg,
			})
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
				runner, err := NewRunner(c.extensions, types.CommandRef{
					Extension: c.alias,
					Command:   msg.Command,
					Params:    msg.Params,
				})
				if err != nil {
					return c, func() tea.Msg {
						return err
					}
				}
				return c, PushPageCmd(runner)
			case types.CommandModeTTY:
				cmd, err := c.extension.Cmd(command.Name, types.CommandInput{Params: msg.Params})
				if err != nil {
					return c, func() tea.Msg {
						return err
					}
				}

				buffer := bytes.Buffer{}
				cmd.Stdout = &buffer
				return c, tea.ExecProcess(cmd, func(err error) tea.Msg {
					if err != nil {
						return err
					}

					if len(buffer.Bytes()) == 0 {
						return nil
					}

					var command types.Command
					if err := json.Unmarshal(buffer.Bytes(), &command); err != nil {
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
		if input.Params != nil {
			c.params = input.Params
		} else {
			input.Params = c.params
		}

		output, err := c.extension.Run(c.command.Name, input)
		if err != nil {
			return err
		}

		if err := schemas.ValidatePage(output); err != nil {
			return err
		}

		var page types.Page
		if err := json.Unmarshal(output, &page); err != nil {
			return err
		}

		if page.Title != "" {
			termOutput.SetWindowTitle(fmt.Sprintf("%s - %s", page.Title, c.extension.Title))
		} else {
			termOutput.SetWindowTitle(fmt.Sprintf("%s - %s", c.command.Title, c.extension.Title))
		}

		switch page.Type {
		case types.PageTypeDetail:
			var detail types.Detail
			if err := json.Unmarshal(output, &detail); err != nil {
				return err
			}

			return detail
		case types.PageTypeList:
			var list types.List
			if err := json.Unmarshal(output, &list); err != nil {
				return err
			}

			return list
		case types.PageTypeForm:
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
