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
	form          *Form
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

	if c.form != nil {
		c.form.SetSize(w, h)
	}

	c.embed.SetSize(w, h)
}

func (c *Runner) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if c.form != nil {
				c.form = nil
				return c, c.embed.Focus()
			}

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
	case types.Action:
		switch msg.Type {
		case types.ActionTypeRun:
			var formItems []FormItem
			for k, v := range msg.Params {
				switch v := v.(type) {
				case types.Text:
					formItems = append(formItems, NewTextItem(k, v))
				case types.TextArea:
					formItems = append(formItems, NewTextArea(k, v))
				case types.Checkbox:
					formItems = append(formItems, NewCheckbox(k, v))
				case types.Select:
					formItems = append(formItems, NewSelect(k, v))
				}
			}

			if len(formItems) > 0 {
				c.form = NewForm(func(values map[string]any) tea.Msg {
					params := make(map[string]any)
					for k, v := range c.input.Params {
						params[k] = v
					}

					for k, v := range values {
						params[k] = v
					}

					return types.Action{
						Title:   msg.Title,
						Type:    types.ActionTypeRun,
						Command: msg.Command,
						Params:  params,
						Exit:    msg.Exit,
						Reload:  msg.Reload,
					}
				}, formItems...)

				c.form.SetSize(c.width, c.height)
				return c, tea.Sequence(c.form.Init(), c.form.Focus())
			}
			c.form = nil

			command, ok := c.extension.Command(msg.Command)
			if !ok {
				c.embed = NewErrorPage(fmt.Errorf("command %s not found", msg.Command))
				c.embed.SetSize(c.width, c.height)
				return c, c.embed.Init()
			}

			switch command.Mode {
			case types.CommandModePage:
				runner := NewRunner(c.extension, command, types.CommandInput{
					Params: msg.Params,
				})

				return c, PushPageCmd(runner)
			case types.CommandModeSilent:
				return c, func() tea.Msg {
					_, err := c.extension.Run(command.Name, types.CommandInput{
						Params: msg.Params,
					})

					if err != nil {
						return err
					}

					if msg.Reload {
						return types.Action{
							Type: types.ActionTypeReload,
						}
					}

					if msg.Exit {
						return ExitMsg{}
					}

					return nil
				}
			}

		case types.ActionTypeCopy:
			return c, func() tea.Msg {
				if err := clipboard.WriteAll(msg.Text); err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			}
		case types.ActionTypeOpen:
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
		case types.ActionTypeReload:
			if c.input.Params == nil {
				c.input.Params = make(map[string]any)
			}

			for k, v := range msg.Params {
				c.input.Params[k] = v
			}

			return c, c.Reload()
		case types.ActionTypeExit:
			return c, ExitCmd
		}
	case error:
		c.embed = NewErrorPage(msg)
		c.embed.SetSize(c.width, c.height)
		return c, c.embed.Init()
	}

	if c.form != nil {
		form, cmd := c.form.Update(msg)
		c.form = form.(*Form)
		return c, cmd
	}

	var cmd tea.Cmd
	c.embed, cmd = c.embed.Update(msg)
	return c, cmd
}

func (c *Runner) View() string {
	if c.form != nil {
		return c.form.View()
	}

	return c.embed.View()
}

func (c *Runner) Reload() tea.Cmd {
	return tea.Sequence(c.SetIsLoading(true), func() tea.Msg {
		output, err := c.extension.Run(c.command.Name, c.input)
		if err != nil {
			return err
		}

		if err := schemas.ValidatePage(output); err != nil {
			return NewErrorPage(err, types.Action{
				Title: "Copy Script Output",
				Type:  types.ActionTypeCopy,
				Text:  string(output),
				Exit:  true,
			})
		}

		var page types.Page
		if err := json.Unmarshal(output, &page); err != nil {
			return NewErrorPage(err, types.Action{
				Title: "Copy Output",
				Type:  types.ActionTypeCopy,
				Text:  string(output),
				Exit:  true,
			})
		}

		if page.Title != "" {
			termenv.DefaultOutput().SetWindowTitle(fmt.Sprintf("%s - %s", page.Title, c.extension.Title))
		} else {
			termenv.DefaultOutput().SetWindowTitle(fmt.Sprintf("%s - %s", c.command.Title, c.extension.Title))
		}

		switch page.Type {
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
