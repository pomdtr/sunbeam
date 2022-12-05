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

type RunContainer struct {
	width, height int
	currentView   string

	extension app.Extension
	params    map[string]any

	list   *List
	detail *Detail

	Script app.Script
}

func NewRunContainer(manifest app.Extension, script app.Script, params map[string]any) *RunContainer {
	return &RunContainer{
		extension: manifest,
		Script:    script,
		params:    params,
	}
}

func (c *RunContainer) Init() tea.Cmd {
	return c.Run()
}

type CommandOutput string

func (c RunContainer) ScriptCmd() tea.Msg {
	command, err := c.Script.Cmd(c.params)
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

func (c *RunContainer) Run() tea.Cmd {
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
	default:
		return NewErrorCmd(fmt.Errorf("unknown script mode: %s", c.Script.Mode))
	}
}

func (c *RunContainer) SetSize(width, height int) {
	c.width, c.height = width, height
	switch c.currentView {
	case "list":
		c.list.SetSize(width, height)
	case "detail":
		c.detail.SetSize(width, height)
	}
}

func (c *RunContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case CommandOutput:
		switch c.currentView {
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
		case "list":
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
		}
	case ReloadPageMsg:
		for key, input := range msg.With {
			c.params[key] = input.Value
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
	}
	return c, cmd
}

func (c *RunContainer) View() string {
	switch c.currentView {
	case "list":
		return c.list.View()
	case "detail":
		return c.detail.View()
	default:
		return ""
	}
}
