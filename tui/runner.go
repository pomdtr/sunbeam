package tui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/app"
)

type CommandRunner struct {
	width, height int
	currentView   string

	extension NamedExtension
	command   NamedCommand

	with    map[string]any
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

func NewCommandRunner(extension NamedExtension, command NamedCommand, with map[string]any) *CommandRunner {
	if with == nil {
		with = map[string]any{}
	}

	return &CommandRunner{
		header:    NewHeader(),
		footer:    NewFooter(extension.Title),
		extension: extension,
		command:   command,
		with:      with,
	}
}

func (c *CommandRunner) Init() tea.Cmd {
	return c.Run()
}

type CommandOutput []byte

func (c CommandRunner) ScriptCmd() tea.Msg {
	commandInput := app.CommandParams{
		With: c.with,
		Env:  c.environ,
	}

	if c.command.Url != "" {
		payload, err := json.Marshal(commandInput)
		if err != nil {
			return err
		}

		res, err := http.Post(c.command.Url, "application/json", bytes.NewReader(payload))
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

	cmd, err := c.command.Cmd(commandInput, c.extension.Root)
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

func (c *CommandRunner) CheckMissingParameters() []FormItem {
	formItems := make([]FormItem, 0)
	for _, input := range c.command.Inputs {
		if _, ok := c.with[input.Name]; ok {
			continue
		}
		formItem := NewFormItem(input)
		formItems = append(formItems, formItem)
	}

	return formItems
}

func (c CommandRunner) Preferences() map[string]app.FormInput {
	preferences := make([]app.FormInput, 0, len(c.extension.Preferences)+len(c.command.Preferences))
	preferences = append(preferences, c.extension.Preferences...)
	preferences = append(preferences, c.command.Preferences...)

	preferenceMap := make(map[string]app.FormInput)
	for _, preference := range preferences {
		preferenceMap[preference.Name] = preference
	}

	return preferenceMap
}

func (c *CommandRunner) checkMissingPreferences() ([]string, []FormItem) {
	envMap := make(map[string]struct{})
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		envMap[pair[0]] = struct{}{}
	}

	environ := make([]string, 0)
	missing := make([]FormItem, 0)
	for name, param := range c.Preferences() {
		// Skip if already set in environment
		if _, ok := envMap[name]; ok {
			continue
		}

		if pref, ok := keyStore.GetPreference(c.extension.Name, c.command.Name, name); ok {
			environ = append(environ, fmt.Sprintf("%s=%s", name, pref.Value))
		} else {
			missing = append(missing, NewFormItem(param))
		}
	}

	return environ, missing
}

func (c *CommandRunner) SetIsloading(isLoading bool) tea.Cmd {
	switch c.currentView {
	case "list":
		return c.list.SetIsLoading(isLoading)
	case "detail":
		return c.detail.SetIsLoading(isLoading)
	case "form":
		return c.form.SetIsLoading(isLoading)
	default:
		return c.header.SetIsLoading(isLoading)
	}
}

func (c *CommandRunner) Run() tea.Cmd {
	environ, missing := c.checkMissingPreferences()
	if len(missing) > 0 {
		c.currentView = "form"
		title := fmt.Sprintf("%s · Preferences", c.extension.Title)
		c.form = NewForm("preferences", title, missing, func(values map[string]any) tea.Cmd {
			preferences := make([]ScriptPreference, 0)
			for _, input := range c.extension.Preferences {
				value, ok := values[input.Name]
				if !ok {
					continue
				}
				preference := ScriptPreference{
					Name:      input.Name,
					Value:     value,
					Extension: c.extension.Name,
				}
				preferences = append(preferences, preference)
			}

			for _, input := range c.command.Preferences {
				value, ok := values[input.Name]
				if !ok {
					continue
				}
				preference := ScriptPreference{
					Name:      input.Name,
					Value:     value,
					Extension: c.extension.Name,
					Command:   c.command.Name,
				}
				preferences = append(preferences, preference)
			}

			err := keyStore.SetPreference(preferences...)
			if err != nil {
				return NewErrorCmd(err)
			}

			return c.Run()
		})

		c.form.SetSize(c.width, c.height)
		return c.form.Init()
	}
	c.environ = environ

	formItems := c.CheckMissingParameters()
	if len(formItems) > 0 {
		c.currentView = "form"

		title := fmt.Sprintf("%s · Params", c.extension.Title)
		c.form = NewForm("params", title, formItems, func(values map[string]any) tea.Cmd {
			for key, value := range values {
				c.with[key] = value
			}
			return c.Run()
		})
		c.form.SetSize(c.width, c.height)
		return c.form.Init()
	}

	return tea.Sequence(c.SetIsloading(true), c.ScriptCmd)
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
			if c.currentView != "" {
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
				c.detail.SetDetail(page.Detail)
				c.detail.SetSize(c.width, c.height)

				return c, c.detail.Init()
			case "list":
				listItems := make([]ListItem, len(page.List.Items))
				for i, scriptItem := range page.List.Items {
					if scriptItem.Id == "" {
						scriptItem.Id = strconv.Itoa(i)
					}

					listItems[i] = ParseScriptItem(scriptItem)
				}

				c.currentView = "list"
				c.list = NewList(page.Title)
				c.list.filter.emptyText = page.List.EmptyText
				c.list.SetItems(listItems)
				c.list.SetSize(c.width, c.height)

				return c, c.list.Init()
			case "form":
				formItems := make([]FormItem, len(page.Form.Inputs))
				for i, input := range page.Form.Inputs {
					formItems[i] = NewFormItem(input)
				}

				c.currentView = "form"
				c.form = NewForm("command", page.Title, formItems, func(values map[string]any) tea.Cmd {
					with := make(map[string]any)
					for key, value := range values {
						with[key] = value
					}

					for key, value := range page.Form.Target.With {
						with[key] = value
					}

					return NewRunCommandCmd(page.Form.Target.Command, with)
				})
				c.form.SetSize(c.width, c.height)

				return c, c.form.Init()
			}

		case "open-url":
			return c, NewOpenUrlCmd(string(msg))
		case "copy-text":
			return c, NewCopyTextCmd(string(msg))
		}

	case RunCommandMsg:
		command, ok := c.extension.Commands[msg.Command]
		if !ok {
			return c, NewErrorCmd(fmt.Errorf("command not found: %s", msg.Command))
		}

		return c, NewPushCmd(NewCommandRunner(c.extension, NamedCommand{
			Name:    msg.Command,
			Command: command,
		}, msg.With))

	case ReloadPageMsg:
		for key, value := range msg.With {
			c.with[key] = value
		}

		return c, c.Run()
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
	default:
		headerView := c.header.View()
		footerView := c.footer.View()
		padding := make([]string, c.height-lipgloss.Height(headerView)-lipgloss.Height(footerView))
		return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), strings.Join(padding, "\n"), c.footer.View())
	}
}
