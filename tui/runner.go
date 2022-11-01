package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/api"
	"github.com/skratchdot/open-golang/open"
)

type RunContainer struct {
	width, height int
	currentView   string

	extensionName string
	scriptName    string
	params        map[string]any

	form   *Form
	list   *List
	detail *Detail

	script api.Script
}

func NewRunContainer(extensionName string, scriptName string, scriptParams map[string]any) *RunContainer {
	params := make(map[string]any)
	for k, v := range scriptParams {
		params[k] = v
	}

	return &RunContainer{
		extensionName: extensionName,
		scriptName:    scriptName,
		params:        params,
	}
}

func (c *RunContainer) Init() tea.Cmd {
	var ok bool
	c.script, ok = api.Sunbeam.GetScript(c.extensionName, c.scriptName)
	if !ok {
		return NewErrorCmd(fmt.Errorf("Script %s not found", c.scriptName))
	}

	missing := c.script.CheckMissingParams(c.params)

	if len(missing) > 0 {
		c.currentView = "form"
		items := make([]FormItem, len(missing))
		for i, param := range missing {
			items[i] = NewFormItem(param)
		}
		c.form = NewForm(items)
		c.form.SetSize(c.width, c.height)
		return c.form.Init()
	}

	return NewSubmitCmd(c.params)
}

type ListOutput []ListItem
type RawOutput string

func (c *RunContainer) Run(input api.ScriptInput) tea.Cmd {
	return func() tea.Msg {
		switch c.script.Output {
		case "list":
			output, err := c.script.Run(input)
			if err != nil {
				return err
			}

			scriptItems, err := api.ParseListItems(output)
			if err != nil {
				return err
			}

			listItems := make([]ListItem, len(scriptItems))
			for i, item := range scriptItems {
				listItems[i] = NewListItem(c.script.Extension, item)
			}
			return ListOutput(listItems)
		case "raw":
			detail, err := c.script.Run(input)
			if err != nil {
				return err
			}
			return RawOutput(detail)
		case "full":
			_, err := c.script.Run(input)
			if err != nil {
				return err
			}
			return tea.Quit()
		case "clipboard":
			output, err := c.script.Run(input)
			if err != nil {
				return err
			}

			err = clipboard.WriteAll(output)
			if err != nil {
				return err
			}
			return tea.Quit()
		case "browser":
			output, err := c.script.Run(input)
			if err != nil {
				return err
			}

			url := strings.Trim(output, " \n")
			err = open.Run(url)
			if err != nil {
				return err
			}
			return tea.Quit()
		default:
			return fmt.Errorf("Unknown output type %s", c.script.Output)
		}
	}
}

func (c *RunContainer) SetSize(width, height int) {
	c.width, c.height = width, height
	switch c.currentView {
	case "form":
		c.form.SetSize(width, height)
	case "list":
		c.list.SetSize(width, height)
	case "detail":
		c.detail.SetSize(width, height)
	}
}

func (c *RunContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case SubmitMsg:
		for k, v := range msg.values {
			c.params[k] = v
		}
		runCmd := c.Run(api.NewScriptInput(c.params))
		if c.script.Output == "list" {
			c.currentView = "list"
			c.list = NewList()
			c.list.SetSize(c.width, c.height)
			return c, tea.Batch(runCmd, c.list.Init())
		} else {
			c.currentView = "detail"
			c.detail = NewDetail()
			c.detail.SetSize(c.width, c.height)
			return c, tea.Batch(runCmd, c.detail.Init())
		}
	case ListOutput:
		c.list.SetItems(msg)
		return c, nil
	case RawOutput:
		err := c.detail.SetContent(string(msg))
		if err != nil {
			return c, NewErrorCmd(fmt.Errorf("Failed to parse script output %s", err))
		}
		return c, nil
	case tea.WindowSizeMsg:
		c.SetSize(msg.Width, msg.Height)
		return c, nil
	}

	var cmd tea.Cmd
	var container Container

	switch c.currentView {
	case "form":
		container, cmd = c.form.Update(msg)
		c.form, _ = container.(*Form)
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
