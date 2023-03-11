package tui

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/utils"
)

type CommandRunner struct {
	width, height int
	currentView   string

	generator Generator
	dir       string

	header Header
	footer Footer

	list   *List
	detail *Detail
}

type Generator func(string) ([]byte, error)

func NewCommandRunner(generator Generator, dir string) *CommandRunner {
	return &CommandRunner{
		header:      NewHeader(),
		footer:      NewFooter("Sunbeam"),
		currentView: "loading",
		generator:   generator,
		dir:         dir,
	}

}
func (c *CommandRunner) Init() tea.Cmd {
	return tea.Batch(c.SetIsloading(true), c.Run())
}

type CommandOutput []byte

func (c *CommandRunner) Run() tea.Cmd {
	var query string
	if c.currentView == "list" {
		query = c.list.Query()
	}

	return func() tea.Msg {
		output, err := c.generator(query)
		if err != nil {
			return err
		}

		return CommandOutput(output)
	}
}

func (c *CommandRunner) SetIsloading(isLoading bool) tea.Cmd {
	switch c.currentView {
	case "list":
		return c.list.SetIsLoading(isLoading)
	case "detail":
		return c.detail.SetIsLoading(isLoading)
	case "loading":
		return c.header.SetIsLoading(isLoading)
	default:
		return nil
	}
}

func (c *CommandRunner) SetSize(width, height int) {
	c.width, c.height = width, height

	c.header.Width = width
	c.footer.Width = width

	switch c.currentView {
	case "list":
		c.list.SetSize(width, height)
	case "detail":
		c.detail.SetSize(width, height)
	}
}

func (runner *CommandRunner) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if runner.currentView != "loading" {
				break
			}
			return runner, func() tea.Msg {
				return PopPageMsg{}
			}
		}
	case CommandOutput:
		runner.SetIsloading(false)
		if err := schemas.Validate(msg); err != nil {
			return runner, func() tea.Msg {
				return fmt.Errorf("invalid response: %s", err)
			}
		}

		var res schemas.Response
		err := json.Unmarshal(msg, &res)
		if err != nil {
			return runner, func() tea.Msg {
				return err
			}
		}

		if res.Title == "" {
			res.Title = "Sunbeam"
		}

		switch res.Type {
		case "detail":
			runner.currentView = "detail"
			actions := make([]Action, len(res.Actions))
			for i, scriptAction := range res.Actions {
				actions[i] = NewAction(scriptAction, runner.dir)
			}
			runner.detail = NewDetail(res.Title, func() string {
				if res.Detail.Content.Text != "" {
					return res.Detail.Content.Text
				}

				cmd := exec.Command(res.Detail.Content.Command, res.Detail.Content.Args...)
				cmd.Dir = runner.dir
				output, err := cmd.Output()
				if err != nil {
					return err.Error()
				}

				return string(output)
			}, actions)

			runner.detail.SetSize(runner.width, runner.height)

			return runner, runner.detail.Init()
		case "list":
			runner.currentView = "list"

			actions := make([]Action, len(res.Actions))
			for i, scriptAction := range res.Actions {
				actions[i] = NewAction(scriptAction, runner.dir)
			}

			if runner.list == nil {
				runner.list = NewList(res.Title, actions)
			} else {
				runner.list.SetTitle(res.Title)
				runner.list.actions = actions
			}

			if res.List.ShowPreview {
				runner.list.ShowPreview = true
			}
			if res.List.GenerateItems {
				runner.list.GenerateItems = true
			}

			listItems := make([]ListItem, len(res.List.Items))
			for i, scriptItem := range res.List.Items {
				scriptItem := scriptItem

				if scriptItem.Id == "" {
					scriptItem.Id = strconv.Itoa(i)
				}
				listItem := ParseScriptItem(scriptItem, runner.dir)
				if scriptItem.Preview != nil {
					listItem.PreviewFunc = func() string {
						if scriptItem.Preview.Text != "" {
							return scriptItem.Preview.Text
						}

						cmd := exec.Command(scriptItem.Preview.Command, scriptItem.Preview.Args...)
						cmd.Dir = runner.dir
						output, err := cmd.Output()
						if err != nil {
							return err.Error()
						}
						return string(output)
					}
				}
				listItems[i] = listItem
			}
			cmd := runner.list.SetItems(listItems)

			runner.list.SetSize(runner.width, runner.height)
			return runner, tea.Sequence(runner.list.Init(), cmd)
		}

	case ReloadPageMsg:
		return runner, tea.Sequence(runner.SetIsloading(true), runner.Run())
	}

	var cmd tea.Cmd
	var container Page

	switch runner.currentView {
	case "list":
		container, cmd = runner.list.Update(msg)
		runner.list, _ = container.(*List)
	case "detail":
		container, cmd = runner.detail.Update(msg)
		runner.detail, _ = container.(*Detail)
	default:
		runner.header, cmd = runner.header.Update(msg)
	}
	return runner, cmd
}

func (c *CommandRunner) View() string {
	switch c.currentView {
	case "list":
		return c.list.View()
	case "detail":
		return c.detail.View()
	case "loading":
		headerView := c.header.View()
		footerView := c.footer.View()
		padding := make([]string, utils.Max(0, c.height-lipgloss.Height(headerView)-lipgloss.Height(footerView)))
		return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), strings.Join(padding, "\n"), c.footer.View())
	default:
		return ""
	}
}
