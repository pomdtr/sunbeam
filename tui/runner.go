package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/api"
)

type RunContainer struct {
	width, height int
	currentView   string

	form    *Form
	loading *Loading
	list    *List
	detail  *Detail
	err     *Detail

	script api.SunbeamScript
	params map[string]string
}

func NewRunContainer(command api.SunbeamScript, params map[string]string) *RunContainer {
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
		case "full":
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
	if c.loading != nil {
		c.loading.SetSize(width, height)
	}
	if c.form != nil {
		c.form.SetSize(width, height)
	}
	if c.list != nil {
		c.list.SetSize(width, height)
	}
	if c.detail != nil {
		c.detail.SetSize(width, height)
	}
	if c.err != nil {
		c.err.SetSize(width, height)
	}
}

func (c *RunContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case SubmitMsg:
		c.currentView = "loading"
		c.loading = NewLoading()
		c.loading.SetSize(c.width, c.height)
		runCmd := c.Run(api.NewScriptInput(msg.values))
		return c, tea.Batch(c.loading.Init(), runCmd)
	case ListOutput:
		if c.list == nil {
			c.currentView = "list"
			c.list = NewList(c.script.Dynamic)
			c.list.SetItems(msg)
			c.list.SetSize(c.width, c.height)
			return c, c.list.Init()
		}
		c.list.SetItems(msg)
	case DetailOutput:
		if c.detail == nil {
			c.currentView = "detail"
			c.detail = NewDetail(c.script.Format, []Action{})
			err := c.detail.SetContent(string(msg))
			if err != nil {
				return c, NewErrorCmd("Failed to parse script output %s", err)
			}
			c.detail.SetSize(c.width, c.height)
			return c, c.detail.Init()
		}
		err := c.detail.SetContent(string(msg))
		if err != nil {
			return c, NewErrorCmd("Failed to parse script output %s", err)
		}
		return c, nil
	case ReloadMsg:
		return c, c.Run(msg.input)
	case QueryUpdateMsg:
		if c.list.dynamic {
			input := api.ScriptInput{
				Query:  msg.query,
				Params: c.params,
			}
			return c, c.Run(input)
		}
	case tea.WindowSizeMsg:
		c.SetSize(msg.Width, msg.Height)
		return c, nil
	case error:
		c.currentView = "error"
		c.err = NewDetail("raw", nil)
		c.err.SetContent(msg.Error())
		c.err.SetSize(c.width, c.height)
		return c, c.err.Init()
	}

	var cmd tea.Cmd
	switch c.currentView {
	case "form":
		c.form, cmd = c.form.Update(msg)
	case "loading":
		c.loading, cmd = c.loading.Update(msg)
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
	case "loading":
		return c.loading.View()
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
