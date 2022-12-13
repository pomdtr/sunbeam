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

	log.Printf("Running command: %s", command.String())

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

func (c *ScriptRunner) CheckMissingParams() ([]FormItem, error) {
	formItems := make([]FormItem, 0)
	for _, param := range c.Script.Params {
		_, ok := c.with[param.Name]
		if ok {
			continue
		}

		if !param.Required {
			continue
		}

		formItem, err := NewFormItem(param)
		if err != nil {
			return nil, err
		}
		formItems = append(formItems, formItem)
	}

	return formItems, nil
}

func (c *ScriptRunner) checkPreferences() error {
	environ := make(map[string]struct{})

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		environ[pair[0]] = struct{}{}
	}

	for _, param := range c.Script.Preferences {
		_, ok := environ[param.Name]
		if ok {
			continue
		}

		if param.Required {
			return fmt.Errorf("missing required pref %s", param.Name)
		}

		if param.Default != nil {
			switch defaultValue := param.Default.(type) {
			case bool:
				os.Setenv(param.Name, strconv.FormatBool(defaultValue))
			case string:
				os.Setenv(param.Name, defaultValue)
			}
		}

	}

	return nil
}

func (c *ScriptRunner) Run() tea.Cmd {
	err := c.checkPreferences()
	if err != nil {
		return NewErrorCmd(err)
	}

	formItems, err := c.CheckMissingParams()
	if err != nil {
		return NewErrorCmd(err)
	}

	if len(formItems) > 0 {
		c.currentView = "form"

		c.form = NewForm(c.extension.Name, formItems)
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
		c.list = NewList(c.Script.Title)
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

		c.detail = NewDetail(c.Script.Title)
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
	}
}

func (c *ScriptRunner) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case CommandOutput:
		switch c.Script.Mode {
		case "detail":
			var detail app.Detail
			json.Unmarshal([]byte(msg), &detail)
			c.detail.SetContent(detail.Preview)
			c.detail.SetIsLoading(false)
			actions := make([]Action, len(detail.Actions))
			for i, action := range detail.Actions {
				actions[i] = NewAction(action)
			}
			c.detail.SetActions(actions...)
			return c, nil
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
		for key, value := range msg.Values {
			input, ok := c.with[key]
			if !ok {
				c.with[key] = app.ScriptInput{Value: value}
				continue
			}

			input.Value = value
			c.with[key] = input
		}
		return c, c.Run()
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
