package tui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/api"
)

type RunContainer struct {
	width, height int
	currentView   string

	manifest api.Manifest
	params   map[string]any

	form   *Form
	list   *List
	detail *Detail

	Script api.Script
}

func NewRunContainer(manifest api.Manifest, page api.Script, scriptParams map[string]any) *RunContainer {
	params := make(map[string]any)
	for k, v := range scriptParams {
		params[k] = v
	}

	return &RunContainer{
		manifest: manifest,
		Script:   page,
		params:   params,
	}
}

func (c *RunContainer) Init() tea.Cmd {
	return c.Run()
}

type ListOutput []ListItem
type FullOutput string

func (c *RunContainer) Run() tea.Cmd {
	missing := c.Script.CheckMissingParams(c.params)

	if len(missing) > 0 {
		var err error
		items := make([]FormItem, len(missing))
		for i, param := range missing {
			items[i], err = NewFormItem(param)
			if err != nil {
				return NewErrorCmd(err)
			}
		}

		c.currentView = "form"
		c.form = NewForm(c.Script.Page.Title, items, func(values map[string]any) tea.Cmd {
			for k, v := range values {
				c.params[k] = v
			}

			return c.Run()
		})
		c.form.SetSize(c.width, c.height)
		return c.form.Init()
	}

	runCmd := func() tea.Msg {

		input := api.CommandInput{
			Params: c.params,
		}
		if c.Script.Page.Mode == "generator" {
			input.Query = c.list.Query()
		}
		command, err := c.Script.Cmd(input)
		log.Println("Running command", command)
		if err != nil {
			return NewErrorCmd(err)
		}
		command.Dir = c.manifest.Dir()
		env := os.Environ()
		command.Env = env

		res, err := command.Output()
		if err != nil {
			var exitErr *exec.ExitError
			if ok := errors.As(err, &exitErr); ok {
				return fmt.Errorf("%s", exitErr.Stderr)
			}
			return err
		}
		output := string(res)

		switch c.Script.Page.Type {
		case "list":
			scriptItems, err := api.ParseListItems(output)
			if err != nil {
				return err
			}

			listItems := make([]ListItem, len(scriptItems))
			for i, scriptItem := range scriptItems {
				scriptItem := scriptItem
				actions := make([]Action, len(scriptItem.Actions))
				for i, scriptAction := range scriptItem.Actions {
					if i == 0 {
						scriptAction.Shortcut = "enter"
					}
					if scriptAction.Extension == "" {
						scriptAction.Extension = c.manifest.Name
					}
					actions[i] = NewAction(scriptAction)
				}
				if scriptItem.Id == "" {
					scriptItem.Id = strconv.Itoa(i)
				}
				listItems[i] = ListItem{
					id:       scriptItem.Id,
					Title:    scriptItem.Title,
					Subtitle: scriptItem.Subtitle,
					Preview:  scriptItem.Preview,
					PreviewCmd: func() string {
						cmd := scriptItem.PreviewCommand()
						if cmd == nil {
							return "No preview command"
						}
						out, err := cmd.Output()
						if err != nil {
							var exitErr *exec.ExitError
							if errors.As(err, &exitErr) {
								return string(exitErr.Stderr)
							}
							return err.Error()
						}

						return string(out)
					},
					Accessories: scriptItem.Accessories,
					Actions:     actions,
				}
			}
			return ListOutput(listItems)
		case "action":
			scriptAction, err := api.ParseAction(output)
			if err != nil {
				return err
			}
			action := NewAction(scriptAction)
			return action
		case "raw":
			return FullOutput(output)
		default:
			return fmt.Errorf("unknown page type %s", c.Script.Page.Type)
		}
	}

	switch c.Script.Page.Type {
	case "list":
		c.currentView = "list"
		if c.list != nil {
			return runCmd
		}
		c.list = NewList(c.Script.Page.Title)
		if c.Script.Page.Mode == "generator" {
			c.list.Dynamic = true
		}
		if c.Script.Page.ShowPreview {
			c.list.ShowPreview = true
		}
		c.list.SetSize(c.width, c.height)
		return tea.Batch(runCmd, c.list.Init())
	case "raw":
		c.currentView = "detail"
		if c.detail != nil {
			return runCmd
		}
		c.detail = NewDetail(c.Script.Page.Title)
		c.detail.SetSize(c.width, c.height)
		return tea.Batch(runCmd, c.detail.Init())
	default:
		return nil
	}
}

func (c *RunContainer) SetSize(width, height int) {
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

func (c *RunContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case ListOutput:
		cmd := c.list.SetItems(msg)
		return c, cmd
	case FullOutput:
		c.detail.SetContent(string(msg))
		c.detail.SetIsLoading(false)
		c.detail.SetActions(
			Action{Title: "Quit", Shortcut: "enter", Cmd: tea.Quit},
			Action{Title: "Copy Output", Shortcut: "ctrl+y", Cmd: func() tea.Msg {
				return CopyTextMsg{
					Text: string(msg),
				}
			}},
			Action{Title: "Reload", Shortcut: "ctrl+r", Cmd: NewReloadCmd(nil)},
		)
		return c, nil
	case ReloadPageMsg:
		for k, v := range msg.Params {
			c.params[k] = v
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

func (c *RunContainer) View() string {
	switch c.currentView {
	case "form":
		return c.form.View()
	case "list":
		return c.list.View()
	case "detail":
		return c.detail.View()
	default:
		return "Unknown view"
	}
}
