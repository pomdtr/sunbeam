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
	params   map[string]any

	list   *List
	detail *Detail

	Page api.Page
}

func NewRunContainer(manifest api.Manifest, page api.Page, scriptParams map[string]any) *RunContainer {
	params := make(map[string]any)
	for k, v := range scriptParams {
		params[k] = v
	}

	return &RunContainer{
		manifest: manifest,
		Page:     page,
		params:   params,
	}
}

func (c *RunContainer) Init() tea.Cmd {
	runCmd := c.Run(c.params)

	if c.Page.Type == "list" {
		c.currentView = "list"
		c.list = NewList(c.Page.Title)
		c.list.SetSize(c.width, c.height)
		return tea.Batch(runCmd, c.list.Init())
	} else {
		c.currentView = "detail"
		c.detail = NewDetail(c.Page.Title)
		c.detail.SetSize(c.width, c.height)
		return tea.Batch(runCmd, c.detail.Init())
	}
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
					}
					if scriptAction.Extension == "" {
						scriptAction.Extension = c.manifest.Name
					}
					actions[i] = NewAction(scriptAction)
				}
				listItems[i] = ListItem{
					Title:       scriptItem.Title,
					Subtitle:    scriptItem.Subtitle,
					Accessories: scriptItem.Accessories,
					Actions:     actions,
				}
			}
			return ListOutput(listItems)
		case "detail":
			return RawOutput(output)
		case "raw":
			return RawOutput(output)
		default:
			return fmt.Errorf("unknown page type %s", c.Page.Type)
		}
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
	case ListOutput:
		c.list.SetItems(msg)
		return c, nil
	case RawOutput:
		err := c.detail.SetContent(string(msg))
		if err != nil {
			return c, NewErrorCmd(fmt.Errorf("failed to parse script output %s", err))
		}
		return c, nil
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
		return "Unknown view"
	}
}
