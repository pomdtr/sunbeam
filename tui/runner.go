package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/api"
)

type RunContainer struct {
	width, height int
	currentView   string

	form   Form
	list   List
	detail Detail
	err    Detail

	script api.SunbeamScript
	params map[string]string
}

func NewRunContainer(command api.SunbeamScript, params map[string]string) *RunContainer {
	if params == nil {
		params = make(map[string]string)
	}
	return &RunContainer{
		script: command,
		params: params,
	}
}

func (c *RunContainer) Init() tea.Cmd {
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
type DetailOutput string

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
			return DetailOutput(detail)
		case "silent":
			_, err := c.script.Run(input)
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
	case "error":
		c.err.SetSize(width, height)
	}
}

func (c *RunContainer) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case SubmitMsg:
		for k, v := range msg.values {
			c.params[k] = v
		}
		runCmd := c.Run(api.NewScriptInput(msg.values))
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
	case DetailOutput:
		err := c.detail.SetContent(string(msg))
		if err != nil {
			return c, NewErrorCmd(fmt.Errorf("Failed to parse script output %s", err))
		}
		return c, nil
	case ReloadMsg:
		for k, v := range msg.Params {
			c.params[k] = v
		}
		return c, c.Run(api.NewScriptInput(c.params))
	case tea.WindowSizeMsg:
		c.SetSize(msg.Width, msg.Height)
		return c, nil
	case error:
		c.currentView = "error"
		errMsg := msg.Error()
		c.err = NewDetail()
		c.err.SetActions(
			Action{Title: "Copy Error", Shortcut: "enter", Msg: CopyMsg{Content: errMsg}},
		)
		c.err.SetContent(errMsg)
		c.err.SetSize(c.width, c.height)
		return c, c.err.Init()
	}

	var cmd tea.Cmd
	switch c.currentView {
	case "form":
		c.form, cmd = c.form.Update(msg)
	case "list":
		c.list, cmd = c.list.Update(msg)
	case "detail":
		c.detail, cmd = c.detail.Update(msg)
	case "error":
		c.err, cmd = c.err.Update(msg)
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
	case "error":
		return c.err.View()
	default:
		return "Unknown view"
	}
}
