package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/api"
)

type RunContainer struct {
	width, height int
	currentView   string

	manifest api.Manifest
	pageName string
	params   map[string]any

	form   *Form
	list   *List
	detail *Detail

	Page api.Page
}

func NewRunContainer(manifest api.Manifest, pageName string, scriptParams map[string]any) *RunContainer {
	params := make(map[string]any)
	for k, v := range scriptParams {
		params[k] = v
	}

	return &RunContainer{
		manifest: manifest,
		pageName: pageName,
		params:   params,
	}
}

func (c *RunContainer) Init() tea.Cmd {
	page, ok := c.manifest.Pages[c.pageName]
	if !ok {
		return NewErrorCmd(fmt.Errorf("Page %s not found", c.pageName))
	}
	c.Page = page
	missing := c.Page.CheckMissingParams(c.params)

	if len(missing) > 0 {
		c.currentView = "form"
		items := make([]FormItem, len(missing))
		for i, param := range missing {
			items[i] = NewFormItem(param)
		}
		c.form = NewForm(c.Page.Title, items)
		c.form.SetSize(c.width, c.height)
		return c.form.Init()
	}

	return NewSubmitCmd(c.params)
}

type ListOutput []ListItem
type RawOutput string

func (c *RunContainer) Run(params map[string]any) tea.Cmd {
	return func() tea.Msg {
		output, err := c.Page.Run(c.manifest.Dir(), params)
		if err != nil {
			return err
		}

		switch c.Page.Type {
		case "list":
			scriptItems, err := api.ParseListItems(output)
			if err != nil {
				return err
			}

			listItems := make([]ListItem, len(scriptItems))
			for i, scriptItem := range scriptItems {
				actions := make([]Action, len(scriptItem.Actions))
				for i, scriptAction := range scriptItem.Actions {
					if i == 0 {
						scriptAction.Shortcut = "enter"
					} else if scriptAction.Shortcut == "" && i < 10 {
						scriptAction.Shortcut = fmt.Sprintf("ctrl+%d", i)
					}
					if scriptAction.Extension == "" {
						scriptAction.Extension = c.manifest.Name
						actions[i] = NewAction(scriptAction)
					}
				}
				listItems[i] = ListItem{
					Title:       scriptItem.Title,
					Subtitle:    scriptItem.Subtitle,
					Accessories: scriptItem.Accessories,
					Actions:     actions,
				}
			}
			return ListOutput(listItems)
		case "raw":
			return RawOutput(output)
		case "full":
			return tea.Quit()
		default:
			return fmt.Errorf("Unknown page type %s", c.Page.Type)
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
		runCmd := c.Run(c.params)
		if c.Page.Type == "list" {
			c.currentView = "list"
			c.list = NewList(c.Page.Title)
			c.list.SetSize(c.width, c.height)
			return c, tea.Batch(runCmd, c.list.Init())
		} else {
			c.currentView = "detail"
			c.detail = NewDetail(c.Page.Title)
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
