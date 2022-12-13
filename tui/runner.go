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
	with      map[string]any

	list   *List
	detail *Detail
	form   *Form

	Script app.Script
}

func NewScriptRunner(manifest app.Extension, script app.Script, params map[string]any) *ScriptRunner {
	return &ScriptRunner{
		extension: manifest,
		Script:    script,
		with:      params,
	}
}

func (c *ScriptRunner) Init() tea.Cmd {
	return c.Run()
}

type CommandOutput string

func (c ScriptRunner) ScriptCmd() tea.Msg {
	commandString, err := c.Script.Cmd(c.with)
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

		if param.Default != nil {
			continue
		}

		switch param.Type {
		case "string", "file", "directory":
			formItem, _ := NewFormItem(param.Name, app.FormItem{
				Type:  "textfield",
				Title: param.Name,
			})
			formItems = append(formItems, formItem)
		case "boolean":
			formItem, _ := NewFormItem(param.Name, app.FormItem{
				Type:  "checkbox",
				Title: param.Name,
			})
			formItems = append(formItems, formItem)
		}
	}

	return formItems, nil
}

func (c *ScriptRunner) Run() tea.Cmd {
	formItems, err := c.CheckMissingParams()
	if err != nil {
		return NewErrorCmd(err)
	}

	if len(formItems) > 0 {
		c.currentView = "form"

		c.form = NewForm(c.extension.Name, formItems, func(formValues map[string]any) tea.Cmd {
			with := make(map[string]any)
			for _, param := range c.Script.Params {
				if value, ok := formValues[param.Name]; ok {
					with[param.Name] = value
				} else if value, ok := c.with[param.Name]; ok {
					with[param.Name] = value
				} else {
					return NewErrorCmd(fmt.Errorf("missing param %s", param.Name))
				}
			}
			c.with = with

			return c.Run()
		})

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
	case ReloadPageMsg:
		for key, value := range msg.With {
			c.with[key] = value
		}
		var cmd tea.Cmd
		if c.currentView == "list" {
			cmd = c.list.header.SetIsLoading(true)
		} else if c.currentView == "detail" {
			cmd = c.detail.header.SetIsLoading(true)
		}
		return c, tea.Batch(cmd, c.Run())
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
