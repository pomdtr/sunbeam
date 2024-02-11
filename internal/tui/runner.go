package tui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/schemas"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
)

type Runner struct {
	embed         Page
	width, height int
	cancel        context.CancelFunc

	extension extensions.Extension
	command   sunbeam.Command
	input     sunbeam.Payload
}

func NewRunner(extension extensions.Extension, input sunbeam.Payload) *Runner {
	var embed Page
	command, ok := extension.Command(input.Command)
	if ok {
		switch command.Mode {
		case sunbeam.CommandModeSearch, sunbeam.CommandModeFilter:
			list := NewList()
			list.SetEmptyText("Loading...")
			if input.Query != "" {
				list.SetQuery(input.Query)
			}

			embed = list
		case sunbeam.CommandModeDetail:
			embed = NewDetail("")
		default:
			embed = NewErrorPage(fmt.Errorf("invalid view type"))
		}
	} else {
		embed = NewErrorPage(fmt.Errorf("command %s not found", input.Command))
	}

	return &Runner{
		embed:     embed,
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
	termenv.DefaultOutput().SetWindowTitle(fmt.Sprintf("%s - %s", c.command.Title, c.extension.Manifest.Title))
	return tea.Batch(c.Reload(), c.embed.Init())
}

func (c *Runner) Focus() tea.Cmd {
	if c.embed == nil {
		return nil
	}
	termenv.DefaultOutput().SetWindowTitle(fmt.Sprintf("%s - %s", c.command.Title, c.extension.Manifest.Title))
	return c.embed.Focus()
}

func (c *Runner) Blur() tea.Cmd {
	c.cancel()
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
			return c, func() tea.Msg {
				manifest, err := extensions.ExtractManifest(c.extension.Entrypoint)
				if err != nil {
					return err
				}
				c.extension.Manifest = manifest

				return ReloadMsg{}
			}
		}
	case ReloadMsg:
		return c, c.Reload()
	case Page:
		c.embed = msg
		c.embed.SetSize(c.width, c.height)
		return c, c.embed.Init()
	case sunbeam.Action:
		var extension extensions.Extension
		if msg.Extension != "" {
			origin, ok := c.extension.Manifest.Imports[msg.Extension]
			if !ok {
				c.embed = NewErrorPage(fmt.Errorf("extension %s not declared as a dependency", msg.Extension))
				return c, c.embed.Init()
			}

			ext, err := extensions.LoadExtension(origin)
			if err != nil {
				c.embed = NewErrorPage(fmt.Errorf("failed to load extension: %w", err))
				return c, c.embed.Init()
			}

			extension = ext
		} else {
			extension = c.extension
		}

		command, ok := extension.Command(msg.Command)
		if !ok {
			c.embed = NewErrorPage(fmt.Errorf("command %s not found", msg.Command))
			c.embed.SetSize(c.width, c.height)
			return c, c.embed.Init()
		}

		ok, missing := extensions.CheckParams(command, msg.Params)
		if !ok {
			c.embed = NewErrorPage(fmt.Errorf("missing required parameters: %v", missing))
			c.embed.SetSize(c.width, c.height)
			return c, c.embed.Init()
		}

		input := sunbeam.Payload{
			Command: msg.Command,
			Params:  make(map[string]any),
		}

		for k, v := range msg.Params {
			input.Params[k] = v
		}

		switch command.Mode {
		case sunbeam.CommandModeSearch, sunbeam.CommandModeFilter, sunbeam.CommandModeDetail:
			runner := NewRunner(extension, input)

			return c, PushPageCmd(runner)
		case sunbeam.CommandModeSilent:
			return c, func() tea.Msg {
				output, err := extension.Output(input)
				if err != nil {
					return PushPageMsg{NewErrorPage(err)}
				}

				if len(output) > 0 {
					var action sunbeam.Action
					if err := json.Unmarshal(output, &action); err != nil {
						return PushPageMsg{NewErrorPage(err)}
					}

					return action
				}

				if msg.Reload {
					return ReloadMsg{}
				}

				return ExitMsg{}
			}
		case sunbeam.CommandModeTTY:
			cmd, err := extension.Cmd(input)
			if err != nil {
				c.embed = NewErrorPage(err)
				c.embed.SetSize(c.width, c.height)
				return c, c.embed.Init()
			}

			return c, tea.ExecProcess(cmd, func(err error) tea.Msg {
				if err != nil {
					return PushPageMsg{NewErrorPage(err)}
				}

				if msg.Reload {
					return ReloadMsg{}
				}

				return ExitMsg{}
			})
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
		if c.cancel != nil {
			c.cancel()
		}

		ctx, cancel := context.WithCancel(context.Background())
		c.cancel = cancel
		defer cancel()

		cmd, err := c.extension.CmdContext(ctx, c.input)
		if err != nil {
			return err
		}

		output, err := cmd.Output()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
			}

			return err
		}

		switch c.command.Mode {
		case sunbeam.CommandModeDetail:
			if err := schemas.ValidateDetail(output); err != nil {
				return err
			}

			var detail sunbeam.Detail
			if err := json.Unmarshal(output, &detail); err != nil {
				return err
			}

			if detail.Markdown != "" {
				page := NewDetail(detail.Markdown, detail.Actions...)
				page.Markdown = true
				return page
			}

			page := NewDetail(detail.Text, detail.Actions...)
			return page
		case sunbeam.CommandModeSearch, sunbeam.CommandModeFilter:
			if err := schemas.ValidateList(output); err != nil {
				return err
			}

			var list sunbeam.List
			if err := json.Unmarshal(output, &list); err != nil {
				return err
			}

			var page *List
			if embed, ok := c.embed.(*List); ok {
				page = embed
				page.SetItems(list.Items...)
				page.SetIsLoading(false)
				page.SetEmptyText(list.EmptyText)
				page.SetActions(list.Actions...)
				page.SetShowDetail(list.ShowDetail)

				if c.command.Mode == sunbeam.CommandModeSearch {
					page.OnQueryChange = func(query string) tea.Cmd {
						c.input.Query = query
						return c.Reload()
					}
					page.ResetSelection()
				}

				return nil
			}

			page = NewList(list.Items...)
			page.SetEmptyText(list.EmptyText)
			page.SetActions(list.Actions...)
			page.SetShowDetail(list.ShowDetail)
			if c.command.Mode == sunbeam.CommandModeSearch {
				page.OnQueryChange = func(query string) tea.Cmd {
					c.input.Query = query
					return c.Reload()
				}
			}

			return page
		default:
			return fmt.Errorf("invalid view type")
		}
	})
}
