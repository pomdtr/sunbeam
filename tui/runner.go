package tui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/utils"
)

type CommandRunner struct {
	width, height int
	currentView   string

	extension NamedExtension
	command   NamedCommand

	with    map[string]app.CommandInput
	environ []string

	header Header
	footer Footer

	list   *List
	detail *Detail
	form   *Form
}

type NamedExtension struct {
	app.Extension
	Name string
}

type NamedCommand struct {
	app.Command
	Name string
}

func NewCommandRunner(extension NamedExtension, command NamedCommand, with map[string]app.CommandInput, environ []string) *CommandRunner {
	runner := CommandRunner{
		header:      NewHeader(),
		footer:      NewFooter(extension.Title),
		extension:   extension,
		currentView: "loading",
		command:     command,
		environ:     environ,
	}

	// Copy the map to avoid modifying the original
	runner.with = make(map[string]app.CommandInput)
	for name, input := range with {
		runner.with[name] = input
	}

	return &runner
}
func (c *CommandRunner) Init() tea.Cmd {
	return c.Run()
}

type CommandOutput []byte

func (c *CommandRunner) Run() tea.Cmd {
	formitems := make([]FormItem, 0)
	for _, param := range c.command.Params {
		input, ok := c.with[param.Name]
		if !ok {
			if param.Default != nil {
				continue
			} else {
				return NewErrorCmd(errors.New("missing required parameter: " + param.Name))
			}
		}

		if input.Value != nil {
			continue
		}

		formitems = append(formitems, NewFormItem(param.Name, input.FormItem))
	}

	// Show form if some parameters are set as input
	if len(formitems) > 0 {
		c.currentView = "form"
		c.form = NewForm(c.extension.Title, formitems)

		c.form.SetSize(c.width, c.height)
		return c.form.Init()
	}

	return tea.Sequence(c.SetIsloading(true), c.Cmd)
}

func (c CommandRunner) Cmd() tea.Msg {
	params := make(map[string]any)
	for name, input := range c.with {
		if input.Value == nil {
			continue
		}
		params[name] = input.Value
	}

	if err := c.command.CheckMissingParams(params); err != nil {
		return err
	}

	commandInput := app.CommandParams{
		With: params,
		Env:  c.environ,
	}

	if c.extension.Root.Scheme != "file" {
		payload, err := json.Marshal(commandInput)
		if err != nil {
			return err
		}

		commandUrl := url.URL{
			Scheme: c.extension.Root.Scheme,
			Host:   c.extension.Root.Host,
			Path:   path.Join(c.extension.Root.Path, c.command.Name),
		}
		res, err := http.Post(commandUrl.String(), "application/json", bytes.NewReader(payload))
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("command failed with status code %d", res.StatusCode)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		if c.command.OnSuccess == "" {
			cmd := exec.Command("cat")
			cmd.Stdin = bytes.NewReader(body)

			return ExecMsg{cmd}
		}

		return CommandOutput(body)
	}

	cmd, err := c.command.Cmd(commandInput, c.extension.Root.Path)
	if err != nil {
		return err
	}

	if c.command.OnSuccess == "" {
		return ExecMsg{cmd}
	}

	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if ok := errors.As(err, &exitErr); ok {
			return fmt.Errorf("command failed with exit code %d, error:\n%s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return err
	}

	return CommandOutput(output)
}

func (c *CommandRunner) SetIsloading(isLoading bool) tea.Cmd {
	switch c.currentView {
	case "list":
		return c.list.SetIsLoading(isLoading)
	case "detail":
		return c.detail.SetIsLoading(isLoading)
	case "form":
		return c.form.SetIsLoading(isLoading)
	case "loading":
		return c.header.SetIsLoading(isLoading)
	default:
		return nil
	}
}

func (c *CommandRunner) SetSize(width, height int) {
	c.width, c.height = width, height

	c.header.Width = width
	c.footer.Width = width

	switch c.currentView {
	case "list":
		c.list.SetSize(width, height)
	case "detail":
		c.detail.SetSize(width, height)
	case "form":
		c.form.SetSize(width, height)
	}
}

func (c *CommandRunner) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if c.currentView != "loading" {
				break
			}
			return c, PopCmd
		}
	case CommandOutput:
		switch c.command.OnSuccess {
		case "push-page":
			var page app.Page
			var v any
			if err := json.Unmarshal(msg, &v); err != nil {
				return c, NewErrorCmd(err)
			}

			if err := app.PageSchema.Validate(v); err != nil {
				return c, NewErrorCmd(err)
			}

			err := json.Unmarshal(msg, &page)
			if err != nil {
				return c, NewErrorCmd(err)
			}

			if page.Title == "" {
				page.Title = c.extension.Title
			}

			switch page.Type {
			case "detail":
				c.currentView = "detail"
				c.detail = NewDetail(page.Title)

				actions := make([]Action, len(page.Detail.Actions))
				for i, scriptAction := range page.Detail.Actions {
					actions[i] = NewAction(scriptAction)
				}
				c.detail.SetActions(actions...)

				if page.Detail.Preview.Text != "" {
					c.detail.viewport.SetContent(page.Detail.Preview.Text)
				}

				if page.Detail.Preview.Command != "" {
					c.detail.PreviewCommand = func() string {
						command, ok := c.extension.Commands[page.Detail.Preview.Command]
						if !ok {
							return ""
						}

						cmd, err := command.Cmd(app.CommandParams{
							With: page.Detail.Preview.With,
							Env:  c.environ,
						}, c.extension.Root.Path)
						if err != nil {
							return err.Error()
						}

						output, err := cmd.Output()
						if err != nil {
							var exitErr *exec.ExitError
							if ok := errors.As(err, &exitErr); ok {
								return fmt.Sprintf("command failed with exit code %d, error:\n%s", exitErr.ExitCode(), exitErr.Stderr)
							}
							return err.Error()
						}

						return string(output)
					}
				}
				c.detail.SetSize(c.width, c.height)

				return c, c.detail.Init()
			case "list":
				c.currentView = "list"
				listItems := make([]ListItem, len(page.List.Items))
				for i, scriptItem := range page.List.Items {
					scriptItem := scriptItem

					if scriptItem.Id == "" {
						scriptItem.Id = strconv.Itoa(i)
					}

					listItem := ParseScriptItem(scriptItem)
					if scriptItem.Preview.Command != "" {
						listItem.PreviewCmd = func() string {
							command, ok := c.extension.Commands[scriptItem.Preview.Command]
							if !ok {
								return fmt.Sprintf("command %s not found", scriptItem.Preview.Command)
							}

							cmd, err := command.Cmd(app.CommandParams{
								With: scriptItem.Preview.With,
								Env:  c.environ,
							}, c.extension.Root.Path)
							if err != nil {
								return err.Error()
							}

							output, err := cmd.Output()
							if err != nil {
								var exitErr *exec.ExitError
								if ok := errors.As(err, &exitErr); ok {
									return fmt.Sprintf("command failed with exit code %d, error:\n%s", exitErr.ExitCode(), exitErr.Stderr)
								}
								return err.Error()
							}

							return string(output)
						}
					}

					listItems[i] = listItem
				}

				c.list = NewList(page.Title)
				c.list.filter.emptyText = page.List.EmptyText
				if page.List.ShowPreview {
					c.list.ShowPreview = true
				}

				c.list.SetItems(listItems)
				c.list.SetSize(c.width, c.height)

				return c, c.list.Init()
			}
		case "open-url":
			return c, NewOpenUrlCmd(string(msg))
		case "copy-text":
			return c, NewCopyTextCmd(string(msg))
		case "reload-page":
			return c, tea.Sequence(PopCmd, NewReloadPageCmd(nil))
		}

	case SubmitFormMsg:
		for key, value := range msg.Values {
			c.with[key] = app.CommandInput{
				Value: value,
			}
		}

		c.currentView = "loading"

		return c, tea.Sequence(c.SetIsloading(true), c.Cmd)

	case RunCommandMsg:
		command, ok := c.extension.Commands[msg.Command]
		if !ok {
			return c, NewErrorCmd(fmt.Errorf("command not found: %s", msg.Command))
		}
		if msg.OnSuccess != "" {
			command.OnSuccess = msg.OnSuccess
		}

		return c, NewPushCmd(NewCommandRunner(c.extension, NamedCommand{
			Name:    msg.Command,
			Command: command,
		}, msg.With, []string{}))

	case ReloadPageMsg:
		for key, value := range msg.With {
			c.with[key] = value
		}

		return c, c.Cmd
	}

	var cmd tea.Cmd
	var container Page

	switch c.currentView {
	case "list":
		container, cmd = c.list.Update(msg)
		c.list, _ = container.(*List)
	case "detail":
		container, cmd = c.detail.Update(msg)
		c.detail, _ = container.(*Detail)
	case "form":
		container, cmd = c.form.Update(msg)
		c.form, _ = container.(*Form)
	default:
		c.header, cmd = c.header.Update(msg)
	}
	return c, cmd
}

func (c *CommandRunner) View() string {
	switch c.currentView {
	case "list":
		return c.list.View()
	case "detail":
		return c.detail.View()
	case "form":
		return c.form.View()
	case "loading":
		headerView := c.header.View()
		footerView := c.footer.View()
		padding := make([]string, utils.Max(0, c.height-lipgloss.Height(headerView)-lipgloss.Height(footerView)))
		return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), strings.Join(padding, "\n"), c.footer.View())
	default:
		return ""
	}
}
