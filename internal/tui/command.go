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

type Command struct {
	embed Page

	width, height int

	extensions Extensions
	ref        types.CommandRef
}

func NewCommand(extensions Extensions, ref types.CommandRef) *Command {
	return &Command{
		extensions: extensions,
		ref:        ref,
	}
}

func (c *Command) Init() tea.Cmd {
	detail := NewDetail("Loading...", "")
	detail.SetSize(c.width, c.height)
	c.embed = detail

	return tea.Batch(c.Run, detail.SetIsLoading(true))
}

func (c *Command) SetSize(w int, h int) {
	c.width = w
	c.height = h

	if c.embed != nil {
		c.embed.SetSize(w, h)
	}
}

func (c *Command) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case types.Form:
		form := msg
		var formitems []FormItem
		for _, item := range form.Inputs {
			formitems = append(formitems, *NewFormItem(item))
		}

		page := NewForm(form.Title, formitems...)
		page.SetSize(c.width, c.height)
		c.embed = page
		return c, c.embed.Init()
	case types.Detail:
		detail := msg

		page := NewDetail(detail.Title, detail.Text, detail.Actions...)
		if detail.Language != "" {
			page.language = detail.Language
		}

		c.embed = page
		c.embed.SetSize(c.width, c.height)
		return c, nil
	case types.List:
		list := msg

		page := NewList(list.Title, list.Items...)
		page.SetSize(c.width, c.height)
		c.embed = page
		return c, c.embed.Init()
	case types.Action:
		action := msg
		return c, func() tea.Msg {
			if action.Type == types.ActionTypeRun && action.Command.Origin == "" {
				action.Command.Origin = c.ref.Origin

				return PushPageMsg{
					Page: NewCommand(c.extensions, action.Command),
				}
			}

			if err := RunAction(action); err != nil {
				return err
			}

			if action.Exit {
				return ExitMsg{}
			}

			return nil
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

func (c *Command) View() string {
	return c.embed.View()
}

func (c *Command) Run() tea.Msg {
	extension, ok := c.extensions[c.ref.Origin]
	if !ok {
		e, err := LoadExtension(c.ref.Origin)
		if err != nil {
			return err
		}

		c.extensions[c.ref.Origin] = e
		extension = e
	}

	command, ok := extension.Command(c.ref.Name)
	if !ok {
		return fmt.Errorf("command %s does not exist", c.ref.Name)
	}

	output, err := extension.Run(command.Name, CommandInput{
		Params: c.ref.Params,
	})
	if err != nil {
		return err
	}

	if command.Mode == types.CommandModeSilent {
		return ExitMsg{}
	}

	if command.Mode == types.CommandModeAction {
		var action types.Action

		if err := json.Unmarshal(output, &action); err != nil {
			return err
		}

		if action.Type == types.ActionTypeRun {
			return fmt.Errorf("cannot chain run actions")
		}

		if err := RunAction(action); err != nil {
			return err
		}

		if action.Exit {
			return ExitMsg{}
		}

		return nil
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
		if err := mapstructure.Decode(page, &detail); err != nil {
			return err
		}

		if detail.Title == "" {
			detail.Title = command.Title
		}

		return detail
	case types.PageTypeList:
		var list types.List
		if err := mapstructure.Decode(page, &list); err != nil {
			return err
		}

		if list.Title == "" {
			list.Title = command.Title
		}

		return list
	case types.PageTypeForm:
		var form types.Form
		if err := mapstructure.Decode(page, &form); err != nil {
			return err
		}

		if form.Title == "" {
			form.Title = command.Title
		}

		return form
	default:
		return fmt.Errorf("invalid command output")
	}
}

func RunAction(action types.Action) error {
	switch action.Type {
	case types.ActionTypeCopy:
		if err := clipboard.WriteAll(action.Text); err != nil {
			return fmt.Errorf("could not copy to clipboard: %s", action.Text)
		}
		return nil
	case types.ActionTypeOpen:
		if err := browser.OpenURL(action.Url); err != nil {
			return fmt.Errorf("could not open url: %s", action.Url)
		}
		return nil
	default:
		return nil
	}
}
