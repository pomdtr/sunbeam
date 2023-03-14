package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/types"
	"github.com/pomdtr/sunbeam/utils"
	"mvdan.cc/sh/v3/shell"
)

type CommandRunner struct {
	width, height int
	currentView   string

	Generator  PageGenerator
	workingDir string

	header Header
	footer Footer

	list   *List
	detail *Detail
	err    *Detail
}

func NewRunner(generator PageGenerator, workingDir string) *CommandRunner {
	return &CommandRunner{
		header:      NewHeader(),
		footer:      NewFooter("Sunbeam"),
		currentView: "loading",
		Generator:   generator,
		workingDir:  workingDir,
	}

}
func (c *CommandRunner) Init() tea.Cmd {
	return tea.Batch(c.SetIsloading(true), c.Refresh())
}

type CommandOutput []byte

func (c *CommandRunner) Refresh() tea.Cmd {
	var query string
	if c.currentView == "list" {
		query = c.list.Query()
	}

	return func() tea.Msg {
		output, err := c.Generator(query)
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
	case "error":
		c.err.SetSize(width, height)
	}
}

func (runner *CommandRunner) Update(msg tea.Msg) (*CommandRunner, tea.Cmd) {
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
		if err := types.Validate(msg); err != nil {
			return runner, func() tea.Msg {
				return fmt.Errorf("invalid response: %s", err)
			}
		}

		var page types.Page
		err := json.Unmarshal(msg, &page)
		if err != nil {
			return runner, func() tea.Msg {
				return err
			}
		}

		if page.Title == "" {
			page.Title = "Sunbeam"
		}

		switch page.Type {
		case types.DetailPage:
			runner.currentView = "detail"

			var detailFunc func() string
			if page.Detail.Text != "" {
				detailFunc = func() string {
					return page.Detail.Text
				}
			} else {
				args, err := shell.Fields(page.Detail.Command, nil)
				if err != nil {
					return runner, func() tea.Msg {
						return fmt.Errorf("invalid command: %s", err)
					}
				}

				extraArgs := []string{}
				if len(args) > 1 {
					extraArgs = args[1:]
				}

				generator := NewCommandGenerator(args[0], extraArgs, runner.workingDir)
				detailFunc = func() string {
					output, err := generator("")
					if err != nil {
						return err.Error()
					}
					return string(output)
				}
			}

			runner.detail = NewDetail(page.Title, detailFunc, page.Actions)
			runner.detail.Language = page.Language
			runner.detail.SetSize(runner.width, runner.height)

			return runner, runner.detail.Init()
		case types.ListPage:
			runner.currentView = "list"

			// Save query string
			var query string
			var selectedId string

			if runner.list != nil {
				query = runner.list.Query()
				if runner.list.Selection() != nil {
					selectedId = runner.list.Selection().Id
				}
			}

			runner.list = NewList(page, runner.workingDir)
			runner.list.SetQuery(query)

			listItems := make([]ListItem, len(page.List.Items))
			for i, scriptItem := range page.List.Items {
				scriptItem := scriptItem
				listItem := ParseScriptItem(scriptItem)
				listItems[i] = listItem
			}

			cmd := runner.list.SetItems(listItems, selectedId)

			runner.list.SetSize(runner.width, runner.height)
			return runner, tea.Sequence(runner.list.Init(), cmd)
		}

	case ReloadPageMsg:
		return runner, tea.Sequence(runner.SetIsloading(true), runner.Refresh())
	case error:
		runner.currentView = "error"
		errorView := NewDetail("Error", msg.Error, []types.Action{
			{
				Type:     types.CopyAction,
				RawTitle: "Copy error",
				Text:     msg.Error(),
			},
		})

		runner.err = errorView
		runner.err.SetSize(runner.width, runner.height)
		return runner, runner.err.Init()
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
	case "error":
		container, cmd = runner.err.Update(msg)
		runner.err, _ = container.(*Detail)
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
	case "error":
		return c.err.View()
	case "loading":
		headerView := c.header.View()
		footerView := c.footer.View()
		padding := make([]string, utils.Max(0, c.height-lipgloss.Height(headerView)-lipgloss.Height(footerView)))
		return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), strings.Join(padding, "\n"), c.footer.View())
	default:
		return ""
	}
}
