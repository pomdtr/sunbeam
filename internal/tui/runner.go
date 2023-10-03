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
	embed Page

	width, height int

	ref CommandRef

	extensions Extensions
}

type ReloadMsg struct {
	params map[string]any
}

type CommandRef struct {
	Script  string
	Command string
	Params  map[string]any
}

func NewRunner(extensions Extensions, ref CommandRef) *Runner {
	detail := NewDetail("")
	detail.SetIsLoading(true)

	return &Runner{
		extensions: extensions,
		ref:        ref,
		embed:      detail,
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
	return tea.Batch(c.Run, c.embed.Init())
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
		case "ctrl+y":
			return c, func() tea.Msg {
				shell := ShellCommand(c.ref)
				if shell == "" {
					return nil
				}

				if err := clipboard.WriteAll(shell); err != nil {
					return err
				}

				return ExitMsg{}
			}
		}
	case types.Form:
		form := msg
		var formitems []FormItem
		for _, item := range form.Items {
			formitems = append(formitems, *NewFormItem(item))
		}

		page := NewForm(c.ref.Command, formitems...)
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
		page.SetSize(c.width, c.height)
		c.embed = page
		return c, tea.Sequence(c.embed.Init(), c.embed.Focus())
	case SubmitMsg:
		return c, func() tea.Msg {
			extension, err := c.extensions.Get(c.ref.Script)
			if err != nil {
				return err
			}

			output, err := extension.Run(
				CommandInput{Command: c.ref.Command, Params: msg, Inputs: msg},
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
		ref := msg
		return c, func() tea.Msg {
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
			case types.CommandTypeRun:
				if msg.Script == "" {
					msg.Script = c.ref.Script
				}

				extension, err := c.extensions.Get(msg.Script)
				if err != nil {
					return err
				}

				command, ok := extension.Command(msg.Command)
				if !ok {
					return fmt.Errorf("command %s not found", msg.Command)
				}

				if command.Mode == types.CommandModeView {
					return PushPageMsg{NewRunner(c.extensions, CommandRef{
						Script:  msg.Script,
						Command: command.Name,
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

				var outputCommand types.Command
				if err := json.Unmarshal(out, &outputCommand); err != nil {
					return err
				}

				return outputCommand
			case types.CommandTypeReload:
				return ReloadMsg{
					params: msg.Params,
				}
			case types.CommandTypeExit:
				return ExitMsg{}
			}

			return PushPageMsg{
				Page: NewRunner(c.extensions, CommandRef{
					Script:  ref.Script,
					Command: ref.Command,
					Params:  ref.Params,
				}),
			}
		}
	case ReloadMsg:
		params := msg.params
		if params != nil {
			c.ref.Params = params
		}
		return c, tea.Sequence(c.SetIsLoading(true), c.Run)
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

func (c *Runner) Run() tea.Msg {
	extension, err := c.extensions.Get(c.ref.Script)
	if err != nil {
		return err
	}

	command, ok := extension.Command(c.ref.Command)
	if !ok {
		return fmt.Errorf("command %s not found", c.ref.Command)
	}

	output, err := extension.Run(
		CommandInput{
			Command: command.Name,
			Params:  c.ref.Params,
		})
	if err != nil {
		return err
	}

	if command.Mode == types.CommandModeNoView {
		if err := schemas.ValidateCommand(output); err != nil {
			return err
		}

		var command types.Command
		if err := json.Unmarshal(output, &command); err != nil {
			return err
		}

		return command
	}

	if err := schemas.ValidatePage(output); err != nil {
		return err
	}

	var page types.Page
	if err := json.Unmarshal(output, &page); err != nil {
		return err
	}

	if page.Title != "" {
		termOutput.SetWindowTitle(fmt.Sprintf("%s - %s", page.Title, extension.Title))
	} else {
		termOutput.SetWindowTitle(fmt.Sprintf("%s - %s", command.Title, extension.Title))
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
}
