package tui

import (
	"encoding/json"
	"fmt"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/browser"
	"github.com/mitchellh/mapstructure"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Runner struct {
	embed Page

	width, height int

	ref CommandRef

	extensions Extensions
}

type CommandRef struct {
	Path    string
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

func (c *Runner) Init() tea.Cmd {
	return tea.Batch(c.Run, c.embed.Init())
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
		return c, c.embed.Init()
	case types.Detail:
		detail := msg

		page := NewDetail(detail.Text, detail.Actions...)
		if detail.Language != "" {
			page.language = detail.Language
		}

		c.embed = page
		c.embed.SetSize(c.width, c.height)
		return c, nil
	case types.List:
		list := msg

		page := NewList(list.Items...)
		page.SetSize(c.width, c.height)
		c.embed = page
		return c, c.embed.Init()
	case SubmitMsg:
		return c, func() tea.Msg {
			extension, err := c.extensions.Get(c.ref.Path)
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
				if err := browser.OpenURL(msg.Url); err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			case types.CommandTypeRun:
				if msg.Origin == "" {
					msg.Origin = c.ref.Path
				}

				if msg.Command == "" {
					return PushPageMsg{NewRunner(c.extensions, CommandRef{
						Path: msg.Origin,
					})}
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
						Path:    msg.Origin,
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
			}

			return PushPageMsg{
				Page: NewRunner(c.extensions, CommandRef{
					Path:    ref.Origin,
					Command: ref.Command,
					Params:  ref.Params,
				}),
			}
		}
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
	extension, err := c.extensions.Get(c.ref.Path)
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
		if len(output) == 0 {
			return ExitMsg{}
		}

		var command types.Command
		if err := json.Unmarshal(output, &command); err != nil {
			return err
		}

		return command
	}

	var page map[string]any
	if err := json.Unmarshal(output, &page); err != nil {
		return err
	}

	rawType, ok := page["type"]
	if !ok {
		return fmt.Errorf("invalid command output")
	}

	pageType, ok := rawType.(string)
	if !ok {
		return fmt.Errorf("invalid command output")
	}

	switch types.PageType(pageType) {
	case types.PageTypeDetail:
		var detail types.Detail
		if detail.Title != "" {
			termOutput.SetWindowTitle(detail.Title)
		} else {
			termOutput.SetWindowTitle(command.Title)
		}

		if err := mapstructure.Decode(page, &detail); err != nil {
			return err
		}

		return detail
	case types.PageTypeList:
		var list types.List
		if list.Title != "" {
			termOutput.SetWindowTitle(list.Title)
		} else {
			termOutput.SetWindowTitle(command.Title)
		}

		if err := mapstructure.Decode(page, &list); err != nil {
			return err
		}

		return list
	case types.PageTypeForm:
		var form types.Form
		if form.Title != "" {
			termOutput.SetWindowTitle(form.Title)
		} else {
			termOutput.SetWindowTitle(command.Title)
		}

		if err := mapstructure.Decode(page, &form); err != nil {
			return err
		}

		return form
	default:
		return fmt.Errorf("invalid command output")
	}
}
