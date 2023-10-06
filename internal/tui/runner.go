package tui

import (
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

	extensions map[string]Extension
	extension  Extension
	alias      string

	input types.CommandInput
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

func NewRunner(extensions map[string]Extension, ref types.CommandRef) *Runner {
	return &Runner{
		extensions: extensions,
		extension:  extensions[ref.Extension],
		alias:      ref.Extension,
		embed:      NewDetail(""),
		input: types.CommandInput{
			Command: ref.Command,
			Params:  ref.Params,
		},
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
	return tea.Batch(c.Run(c.input), c.embed.Init())
}

func (c *Runner) Focus() tea.Cmd {
	return nil
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

		page := NewDetail(detail.Text, detail.Actions...)
		if detail.Language != "" {
			page.language = detail.Language
		}

		c.embed = page
		c.embed.SetSize(c.width, c.height)
		return c, tea.Sequence(c.embed.Init(), c.embed.Focus())
	case types.List:
		list := msg
		page := NewList(list.Items...)
		if list.Reload {
			page.OnQueryChange = func(query string) tea.Cmd {
				input := c.input
				input.Query = query
				return c.Run(input)
			}
		}

		page.SetSize(c.width, c.height)
		if c.embed == nil {
			c.embed = page
			return c, tea.Sequence(c.embed.Init(), c.embed.Focus())
		}

		if list, ok := c.embed.(*List); ok {
			page.SetQuery(list.Query())
		}

		c.embed = page
		return c, tea.Sequence(c.embed.Init(), c.embed.Focus())
	case SubmitMsg:
		return c, func() tea.Msg {
			output, err := c.extension.Run(
				types.CommandInput{Command: c.input.Command, Params: msg, Inputs: msg},
			)
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
						types.CommandInput{
							Command: command.Name,
							Params:  msg.Params,
						})
					if err != nil {
						return err
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
				runner := NewRunner(c.extensions, types.CommandRef{
					Extension: c.alias,
					Command:   msg.Command,
					Params:    msg.Params,
				})
				return c, PushPageCmd(runner)
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
			input := c.input
			if msg.Params != nil {
				input.Params = msg.Params
			}

			return c, c.Run(input)
		case types.CommandTypeExit:
			return c, ExitCmd
		case types.CommandTypePop:
			if msg.Reload {
				return c, tea.Sequence(PopPageCmd, ReloadCmd(nil))
			}
			return c, PopPageCmd
		case types.CommandTypePass:
			return c, nil
		}
	case ReloadMsg:
		if msg.params != nil {
			c.input.Params = msg.params
		}

		return c, tea.Sequence(c.SetIsLoading(true))
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

func IsRootCommand(command types.CommandSpec) bool {
	if command.Hidden {
		return false
	}

	for _, param := range command.Params {
		if !param.Optional {
			return false
		}
	}

	return true
}

func (c *Runner) View() string {
	return c.embed.View()
}

func (c *Runner) Run(input types.CommandInput) tea.Cmd {
	c.input = input
	return tea.Sequence(c.SetIsLoading(true), func() tea.Msg {
		output, err := c.extension.Run(input)
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
			termOutput.SetWindowTitle(c.extension.Title)
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
