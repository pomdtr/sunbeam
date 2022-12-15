package tui

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/app"
)

type ScriptRunner struct {
	width, height int
	currentView   string

	extension app.Extension
	with      app.ScriptInputs
	environ   []string

	list   *List
	detail *Detail
	form   *Form

	Script app.Script
}

func NewScriptRunner(manifest app.Extension, script app.Script, params app.ScriptInputs) *ScriptRunner {
	with := make(app.ScriptInputs)
	for key, value := range params {
		with[key] = value
	}
	return &ScriptRunner{
		extension: manifest,
		Script:    script,
		with:      with,
	}
}

func (c *ScriptRunner) Init() tea.Cmd {
	return c.Run()
}

type CommandOutput string

func (c ScriptRunner) ScriptCmd() tea.Msg {
	with := make(map[string]any)
	for key, input := range c.with {
		with[key] = input.Value
	}
	commandString, err := c.Script.Cmd(with)
	command := exec.Command("sh", "-c", commandString)

	if c.Script.Mode == "generator" {
		command.Stdin = strings.NewReader(c.list.Query())
	}
	if err != nil {
		return err
	}

	switch c.Script.Cwd {
	case "homeDir":
		command.Dir, err = os.UserHomeDir()
		if err != nil {
			return err
		}
	case "extensionDir":
		command.Dir = c.extension.Dir()
	case "currentDir":
		command.Dir, err = os.Getwd()
		if err != nil {
			return err
		}
	default:
		command.Dir = c.extension.Dir()
	}

	command.Env = os.Environ()
	command.Env = append(command.Env, c.environ...)

	log.Printf("Running command: %s, env: %s", command.String(), c.environ)

	res, err := command.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if ok := errors.As(err, &exitErr); ok {
			return fmt.Errorf("command failed with exit code %d, error: %s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return err
	}

	return CommandOutput(string(res))
}

func (c *ScriptRunner) CheckMissingParameters() []FormItem {
	formItems := make([]FormItem, 0)
	for _, param := range c.Script.Inputs {
		_, ok := c.with[param.Name]
		if ok {
			continue
		}

		if !param.Required {
			continue
		}

		formItem := NewFormInput(param)
		formItems = append(formItems, formItem)
	}

	return formItems
}

func (c ScriptRunner) Preferences() []app.ScriptInput {
	preferences := make([]app.ScriptInput, 0, len(c.extension.Preferences)+len(c.Script.Preferences))
	preferences = append(preferences, c.extension.Preferences...)
	preferences = append(preferences, c.Script.Preferences...)

	return preferences
}

func (c *ScriptRunner) checkPreferences(preferences []app.ScriptPreference) (environ []string, missing []FormItem) {
	preferenceMap := make(map[string]any)
	for _, preference := range preferences {
		if preference.Extension != c.extension.Name {
			continue
		}

		if preference.Script != "" && preference.Script != c.Script.Name {
			continue
		}

		preferenceMap[preference.Name] = preference.Value
	}

	envMap := make(map[string]struct{})
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		envMap[pair[0]] = struct{}{}
	}

	for _, param := range c.Preferences() {
		if pref, ok := envMap[param.Name]; ok {
			environ = append(environ, fmt.Sprintf("%s=%s", param.Name, pref))
			continue
		}

		if pref, ok := preferenceMap[param.Name]; ok {
			environ = append(environ, fmt.Sprintf("%s=%s", param.Name, pref))
			continue
		}

		if param.Required {
			missing = append(missing, FormItem{
				Title:     param.Title,
				Id:        param.Name,
				FormInput: NewFormInput(param),
			})
			continue
		}

		if param.Default != "" {
			environ = append(environ, param.Name, fmt.Sprintf("%s=%s", param.Name, param.Default))
		}

	}

	return environ, missing
}

func (c *ScriptRunner) Run() tea.Cmd {
	preferences, err := getPreferences()
	if err != nil {
		return NewErrorCmd(err)
	}

	environ, missing := c.checkPreferences(preferences)
	if len(missing) > 0 {
		c.currentView = "form"
		title := fmt.Sprintf("%s · Preferences", c.extension.Title)
		c.form = NewForm("preferences", title, missing)
		c.form.SetSize(c.width, c.height)
		return c.form.Init()
	}

	c.environ = environ
	formItems := c.CheckMissingParameters()

	if len(formItems) > 0 {
		c.currentView = "form"

		title := fmt.Sprintf("%s · Params", c.extension.Title)
		c.form = NewForm("params", title, formItems)
		c.form.SetSize(c.width, c.height)
		return c.form.Init()
	}

	switch c.Script.Mode {
	case "filter", "generator":
		c.currentView = "list"
		if c.list != nil {
			cmd := c.list.SetIsLoading(true)
			return tea.Batch(cmd, c.ScriptCmd)
		}
		c.list = NewList(c.extension.Title)
		if c.Script.Mode == "generator" {
			c.list.Dynamic = true
		}
		if c.Script.ShowPreview {
			c.list.ShowPreview = true
		}
		c.list.SetSize(c.width, c.height)
		cmd := c.list.SetIsLoading(true)
		return tea.Batch(c.ScriptCmd, c.list.Init(), cmd)
	case "detail":
		c.currentView = "detail"
		if c.detail != nil {
			cmd := c.detail.SetIsLoading(true)
			return tea.Batch(cmd, c.ScriptCmd)
		}

		c.detail = NewDetail(c.extension.Title)
		c.detail.SetSize(c.width, c.height)
		cmd := c.detail.SetIsLoading(true)
		return tea.Batch(c.ScriptCmd, cmd, c.detail.Init())
	case "snippet", "quicklink":
		return c.ScriptCmd
	default:
		return NewErrorCmd(fmt.Errorf("unknown script mode: %s", c.Script.Mode))
	}
}

func (c *ScriptRunner) SetSize(width, height int) {
	c.width, c.height = width, height
	switch c.currentView {
	case "list":
		c.list.SetSize(width, height)
	case "detail":
		c.detail.SetSize(width, height)
	case "form":
		c.form.SetSize(width, height)
	}
}

func (c *ScriptRunner) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case CommandOutput:
		switch c.Script.Mode {
		case "detail":
			var detail app.Detail
			err := json.Unmarshal([]byte(msg), &detail)
			if err != nil {
				return c, NewErrorCmd(err)
			}

			c.detail.SetIsLoading(false)
			cmd := c.detail.SetDetail(detail)

			return c, cmd
		case "filter", "generator":
			scriptItems, err := app.ParseListItems(string(msg))
			if err != nil {
				return c, NewErrorCmd(err)
			}
			listItems := make([]ListItem, len(scriptItems))

			for i, scriptItem := range scriptItems {
				if scriptItem.Id == "" {
					scriptItem.Id = strconv.Itoa(i)
				}

				for i, action := range scriptItem.Actions {
					if action.Extension == "" {
						action.Extension = c.extension.Name
					}
					scriptItem.Actions[i] = action
				}

				listItems[i] = ParseScriptItem(scriptItem)
			}

			cmd := c.list.SetItems(listItems)
			c.list.SetIsLoading(false)
			return c, cmd
		case "command":
			return c, tea.Quit
		case "quicklink":
			return c, NewOpenUrlCmd(string(msg))
		case "snippet":
			return c, NewCopyTextCmd(string(msg))
		}
	case SubmitMsg:
		switch msg.Name {
		case "preferences":
			preferences := make([]app.ScriptPreference, 0)
			for _, input := range c.extension.Preferences {
				value, ok := msg.Values[input.Name]
				if !ok {
					continue
				}
				preference := app.ScriptPreference{
					Name:      input.Name,
					Value:     value,
					Extension: c.extension.Name,
				}
				preferences = append(preferences, preference)
			}

			for _, input := range c.Script.Preferences {
				value, ok := msg.Values[input.Name]
				if !ok {
					continue
				}
				preference := app.ScriptPreference{
					Name:      input.Name,
					Value:     value,
					Extension: c.extension.Name,
					Script:    c.Script.Name,
				}
				preferences = append(preferences, preference)
			}

			err := setPreference(preferences...)
			if err != nil {
				return c, NewErrorCmd(err)
			}

			return c, c.Run()
		case "params":
			for key, value := range msg.Values {
				input, ok := c.with[key]
				if !ok {
					c.with[key] = app.ScriptParam{Value: value}
					continue
				}

				input.Value = value
				c.with[key] = input
			}
			return c, c.Run()
		}

	case ReloadPageMsg:
		for key, value := range msg.With {
			c.with[key] = value
		}

		return c, c.Run()
	}

	var cmd tea.Cmd
	var container Container

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
	}
	return c, cmd
}

func (c *ScriptRunner) View() string {
	switch c.currentView {
	case "list":
		return c.list.View()
	case "detail":
		return c.detail.View()
	case "form":
		return c.form.View()
	default:
		return ""
	}
}
