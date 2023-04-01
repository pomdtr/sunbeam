package internal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/browser"
	"github.com/pomdtr/sunbeam/types"
	"github.com/pomdtr/sunbeam/utils"
)

type CommandRunner struct {
	width, height int
	currentView   RunnerView

	values map[string]string

	Generator PageGenerator

	header Header
	footer Footer

	form   *Form
	list   *List
	detail *Detail
	err    *Detail
}

type RunnerView int

const (
	RunnerViewList RunnerView = iota
	RunnerViewDetail
	RunnerViewLoading
)

type PageValidator func([]byte) error

func NewRunner(generator PageGenerator) *CommandRunner {
	return &CommandRunner{
		header:      NewHeader(),
		footer:      NewFooter("Sunbeam"),
		currentView: RunnerViewLoading,
		Generator:   generator,
	}

}
func (c *CommandRunner) Init() tea.Cmd {
	return tea.Batch(c.SetIsloading(true), c.Refresh)
}

type CommandOutput []byte

func (c *CommandRunner) Refresh() tea.Msg {
	var query string
	if c.currentView == RunnerViewList {
		query = c.list.Query()
	}

	output, err := c.Generator(query)
	if err != nil {
		return err
	}

	return CommandOutput(output)
}

func (runner *CommandRunner) handleAction(action types.Action) tea.Cmd {
	action = ExpandAction(action, runner.values)
	switch action.Type {
	case types.ReloadAction:
		return tea.Sequence(runner.SetIsloading(true), runner.Refresh)
	case types.OpenAction:
		var target string
		if action.Url != "" {
			target = action.Url
		} else if action.Path != "" {
			target = action.Path
		}

		if err := browser.OpenURL(target); err != nil {
			return func() tea.Msg {
				return err
			}
		}

		return tea.Quit
	case types.CopyAction:
		err := clipboard.WriteAll(action.Text)
		if err != nil {
			return func() tea.Msg {
				return fmt.Errorf("failed to copy text to clipboard: %s", err)
			}
		}

		return tea.Quit
	case types.ReadAction, types.FetchAction:
		return func() tea.Msg {
			generator := NewFileGenerator(action.Path)
			return PushPageMsg{
				runner: NewRunner(generator),
			}
		}

	case types.RunAction:
		switch action.OnSuccess {
		case types.PushOnSuccess:
			generator := NewCommandGenerator(action.Command, action.Dir)
			return func() tea.Msg {
				return PushPageMsg{NewRunner(generator)}
			}
		case types.ReloadOnSuccess:
			return func() tea.Msg {
				_, err := utils.RunCommand(action.Command, action.Dir)
				if err != nil {
					return err
				}

				return types.Action{
					Type: types.ReloadAction,
				}
			}
		case types.ExitOnSuccess:
			return func() tea.Msg {
				_, err := utils.RunCommand(action.Command, action.Dir)
				if err != nil {
					return err
				}

				return tea.Quit
			}

		default:
			return func() tea.Msg {
				return fmt.Errorf("unsupported onSuccess")
			}
		}
	default:
		return func() tea.Msg {
			return fmt.Errorf("unknown action type")
		}
	}
}

func (c *CommandRunner) SetIsloading(isLoading bool) tea.Cmd {
	switch c.currentView {
	case RunnerViewList:
		return c.list.SetIsLoading(isLoading)
	case RunnerViewDetail:
		return c.detail.SetIsLoading(isLoading)
	case RunnerViewLoading:
		return c.header.SetIsLoading(isLoading)
	default:
		return nil
	}
}

func (c *CommandRunner) SetSize(width, height int) {
	c.width, c.height = width, height

	c.header.Width = width
	c.footer.Width = width

	if c.form != nil {
		c.form.SetSize(width, height)
		return
	}

	if c.err != nil {
		c.err.SetSize(width, height)
		return
	}

	switch c.currentView {
	case RunnerViewList:
		c.list.SetSize(width, height)
	case RunnerViewDetail:
		c.detail.SetSize(width, height)
	}
}

func (runner *CommandRunner) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if runner.form != nil {
				runner.form = nil
				return runner, nil
			}

			if runner.currentView == RunnerViewLoading {
				return runner, func() tea.Msg {
					return PopPageMsg{}
				}
			}
		}
	case CommandOutput:
		runner.SetIsloading(false)

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
			var detailFunc func() string
			if page.Text != "" {
				detailFunc = func() string {
					return page.Text
				}
			} else if page.Command != "" {
				detailFunc = func() string {
					output, err := utils.RunCommand(page.Command, page.Dir)
					if err != nil {
						return err.Error()
					}
					return string(output)
				}
			} else {
				return runner, func() tea.Msg {
					return fmt.Errorf("detail page must have either text or command")
				}
			}

			runner.currentView = RunnerViewDetail
			runner.detail = NewDetail(page.Title, detailFunc, page.Actions)
			runner.detail.Language = page.Language
			runner.detail.SetSize(runner.width, runner.height)

			return runner, runner.detail.Init()
		case types.ListPage:
			runner.currentView = RunnerViewList

			// Save query string
			var query string
			var selectedId string

			if runner.list != nil {
				query = runner.list.Query()
				if runner.list.Selection() != nil {
					selectedId = runner.list.Selection().ID()
				}
			}

			runner.list = NewList(page)
			runner.list.SetQuery(query)

			listItems := make([]ListItem, len(page.Items))
			for i, scriptItem := range page.Items {
				scriptItem := scriptItem
				listItem := ParseScriptItem(scriptItem)
				listItems[i] = listItem
			}

			cmd := runner.list.SetItems(listItems, selectedId)

			runner.list.SetSize(runner.width, runner.height)
			return runner, tea.Sequence(runner.list.Init(), cmd)
		}

	case types.Action:
		if len(msg.Inputs) > len(runner.values) {
			formItems := make([]FormItem, len(msg.Inputs))
			for i, input := range msg.Inputs {
				item, err := NewFormItem(input)
				if err != nil {
					return runner, func() tea.Msg {
						return fmt.Errorf("failed to create form input: %s", err)
					}
				}
				formItems[i] = item
			}

			form := NewForm(formItems, func(values map[string]string) tea.Msg {
				runner.values = values
				return msg
			})

			runner.form = form
			runner.SetSize(runner.width, runner.height)
			return runner, form.Init()
		}

		runner.form = nil
		cmd := runner.handleAction(msg)
		return runner, cmd
	case error:
		errorView := NewDetail("Error", msg.Error, []types.Action{
			{
				Type:  types.CopyAction,
				Title: "Copy error",
				Text:  msg.Error(),
			},
		})

		runner.err = errorView
		runner.err.SetSize(runner.width, runner.height)
		return runner, runner.err.Init()
	}

	var cmd tea.Cmd
	var container Page

	if runner.form != nil {
		container, cmd = runner.form.Update(msg)
		runner.form, _ = container.(*Form)
		return runner, cmd
	}

	if runner.err != nil {
		container, cmd = runner.err.Update(msg)
		runner.err, _ = container.(*Detail)
		return runner, cmd
	}

	switch runner.currentView {
	case RunnerViewList:
		container, cmd = runner.list.Update(msg)
		runner.list, _ = container.(*List)
	case RunnerViewDetail:
		container, cmd = runner.detail.Update(msg)
		runner.detail, _ = container.(*Detail)
	default:
		runner.header, cmd = runner.header.Update(msg)
	}
	return runner, cmd
}

func (c *CommandRunner) View() string {
	if c.form != nil {
		return c.form.View()
	}

	if c.err != nil {
		return c.err.View()
	}

	switch c.currentView {
	case RunnerViewList:
		return c.list.View()
	case RunnerViewDetail:
		return c.detail.View()
	case RunnerViewLoading:
		headerView := c.header.View()
		footerView := c.footer.View()
		padding := make([]string, utils.Max(0, c.height-lipgloss.Height(headerView)-lipgloss.Height(footerView)))
		return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), strings.Join(padding, "\n"), c.footer.View())
	default:
		return ""
	}
}
