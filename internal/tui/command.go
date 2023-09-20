package tui

import (
	"encoding/json"
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/browser"
	"github.com/mitchellh/mapstructure"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Command struct {
	header   Header
	viewport viewport.Model
	embed    Page
	footer   Footer

	width, height int

	commandRef types.CommandRef

	extensions Extensions
	command    types.Command
}

func NewCommand(extensions Extensions, commandRef types.CommandRef) *Command {
	viewport := viewport.New(0, 0)
	return &Command{
		header:   NewHeader(),
		footer:   NewFooter("Loading..."),
		viewport: viewport,

		extensions: extensions,
		commandRef: commandRef,
	}
}

func (c *Command) Init() tea.Cmd {
	extension, ok := c.extensions[c.commandRef.Extension]
	if !ok {
		return nil
	}

	command, ok := extension.Commands[c.commandRef.Name]
	if !ok {
		return nil
	}
	c.command = command

	if c.command.Mode == "" {
		c.embed = NewDetail(c.command.Title, "")
		c.embed.SetSize(c.width, c.height)
		return tea.Sequence(c.embed.Init())
	}

	return tea.Batch(FocusCmd, c.header.SetIsLoading(true), c.Reload)
}

func (c *Command) SetSize(w int, h int) {
	c.width = w
	c.height = h

	c.header.Width = w
	c.footer.Width = w
	c.viewport.Width = w
	c.viewport.Height = h - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())

	if c.embed != nil {
		c.embed.SetSize(w, h)
	}
}

func (c *Command) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case types.Form:
		form := msg
		if form.Title == "" {
			form.Title = c.command.Title
		}

		page, ok := c.embed.(*Form)
		if !ok {
			return c, nil
		}

		var formitems []FormItem
		for _, item := range form.Items {
			formitems = append(formitems, *NewFormItem(item))
		}
		page.SetItems(formitems...)

		c.embed = page
		return c, c.embed.Init()

	case types.Detail:
		detail := msg
		if detail.Title == "" {
			detail.Title = c.command.Title
		}

		page := NewDetail(detail.Title, detail.Text, detail.Actions...)
		if detail.Language != "" {
			page.language = detail.Language
		}

		c.embed = page
		c.embed.SetSize(c.width, c.height)
		return c, nil
	case types.List:
		list := msg
		if list.Title == "" {
			list.Title = c.command.Title
		}

		page := NewList(list.Title, list.Items...)
		page.SetSize(c.width, c.height)
		c.embed = page
		return c, c.embed.Init()
	case types.Action:
		action := msg
		return c, func() tea.Msg {
			if action.Type == types.ActionTypeRun && action.Command.Extension == "" {
				action.Command.Extension = c.commandRef.Extension
			}

			return RunAction(c.extensions, action)
		}
	case error:
		c.embed = NewErrorPage(msg)
		c.embed.SetSize(c.width, c.height)
		return c, c.embed.Init()
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	if c.embed != nil {
		c.embed, cmd = c.embed.Update(msg)
		cmds = append(cmds, cmd)
	}

	c.header, cmd = c.header.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *Command) View() string {
	if c.embed != nil {
		return c.embed.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}

func (d *Command) Reload() tea.Msg {
	extension, ok := d.extensions[d.commandRef.Extension]
	if !ok {
		return fmt.Errorf("extension %s does not exist", d.commandRef.Extension)
	}
	output, err := extension.Run(d.commandRef.Name, CommandInput{
		Params: d.commandRef.Params,
	})
	if err != nil {
		return err
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

		return detail
	case types.PageTypeList:
		var list types.List
		if err := mapstructure.Decode(page, &list); err != nil {
			return err
		}

		return list
	case types.PageTypeForm:
		var form types.Form
		if err := mapstructure.Decode(page, &form); err != nil {
			return err
		}

		return form
	default:
		return fmt.Errorf("invalid command output")
	}
}

func RunAction(extensions Extensions, action types.Action) tea.Msg {
	switch action.Type {
	case types.ActionTypeCopy:
		if err := clipboard.WriteAll(action.Text); err != nil {
			return fmt.Errorf("could not copy to clipboard: %s", action.Text)
		}

		if action.Exit {
			return ExitMsg{}
		}

		return nil
	case types.ActionTypeOpen:
		if err := browser.OpenURL(action.Url); err != nil {
			return fmt.Errorf("could not open url: %s", action.Url)
		}

		if action.Exit {
			return ExitMsg{}
		}

		return nil
	case types.ActionTypeReload:
		return nil
	case types.ActionTypeRun:
		extension, ok := extensions[action.Command.Extension]
		if !ok {
			return fmt.Errorf("extension %s does not exist", action.Command.Extension)
		}
		command, ok := extension.Commands[action.Command.Name]
		if !ok {
			return fmt.Errorf("command %s does not exist", action.Command.Name)
		}

		switch command.Mode {
		case types.CommandModePage:
			return PushPageMsg{Page: NewCommand(extensions, action.Command)}
		case types.CommandModeSilent:

			if _, err := extension.Run(action.Command.Name, CommandInput{
				Params: action.Command.Params,
			}); err != nil {
				return err
			}

			return ExitMsg{}
		case types.CommandModeAction:
			output, err := extension.Run(action.Command.Name, CommandInput{
				Params: action.Command.Params,
			})
			if err != nil {
				return err
			}
			var action types.Action
			if err := json.Unmarshal(output, &action); err != nil {
				return err
			}

			if action.Type == types.ActionTypeRun {
				command, ok := extension.Commands[action.Command.Name]
				if !ok {
					return fmt.Errorf("command %s does not exist", action.Command.Name)
				}

				if command.Mode == types.CommandModeAction {
					return fmt.Errorf("too many nested actions")
				}

				return PushPageMsg{Page: NewCommand(extensions, action.Command)}
			}

			return RunAction(extensions, action)
		}

	}

	return nil
}
