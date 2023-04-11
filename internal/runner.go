package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/pkg/browser"
	"github.com/pomdtr/sunbeam/types"
	"github.com/pomdtr/sunbeam/utils"
)

type CommandRunner struct {
	width, height int
	currentView   RunnerView

	Generator PageGenerator

	header Header
	footer Footer

	currentPage *types.Page

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
	RunnerViewForm
)

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

func (runner *CommandRunner) Refresh() tea.Msg {
	page, err := runner.Generator()
	if err != nil {
		return err
	}

	return page
}

func (runner *CommandRunner) handleAction(action types.Action) tea.Cmd {
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		action = ExpandAction(action, fmt.Sprintf("${env:%s}", pair[0]), pair[1])
	}

	switch action.Type {
	case types.ReloadAction:
		return tea.Sequence(runner.SetIsloading(true), runner.Refresh)
	case types.OpenPathAction:
		return func() tea.Msg {
			if err := browser.OpenURL(action.Path); err != nil {
				return func() tea.Msg {
					return err
				}
			}
			return tea.Quit()
		}
	case types.OpenUrlAction:
		return func() tea.Msg {
			if err := browser.OpenURL(action.Url); err != nil {
				return func() tea.Msg {
					return err
				}
			}

			return tea.Quit()
		}
	case types.CopyAction:
		return func() tea.Msg {
			err := clipboard.WriteAll(action.Text)
			if err != nil {
				return func() tea.Msg {
					return fmt.Errorf("failed to copy text to clipboard: %s", err)
				}
			}
			return tea.Quit()
		}

	case types.PushPageAction:
		return func() tea.Msg {
			if action.Page == nil {
				return fmt.Errorf("page is nil")
			}
			var generator PageGenerator
			if action.Page.Type == types.StaticTarget {
				generator = NewFileGenerator(action.Page.Path)
			} else if action.Page.Type == types.DynamicTarget {
				generator = NewCommandGenerator(action.Page.Command, action.Page.Input, action.Page.Dir)
			} else {
				return fmt.Errorf("unknown page type")
			}
			return PushPageMsg{
				runner: NewRunner(generator),
			}
		}

	case types.RunAction:
		switch action.OnSuccess {
		case types.ReloadOnSuccess:
			return func() tea.Msg {
				_, err := utils.RunCommand(action.Command, strings.NewReader(action.Input), action.Dir)
				if err != nil {
					return err
				}

				return types.Action{
					Type: types.ReloadAction,
				}
			}
		case types.ExitOnSuccess:
			return func() tea.Msg {
				_, err := utils.RunCommand(action.Command, strings.NewReader(action.Input), action.Dir)
				if err != nil {
					return err
				}

				return tea.Quit()
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
	case RunnerViewForm:
		return c.form.SetIsLoading(isLoading)
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

	if c.err != nil {
		c.err.SetSize(width, height)
		return
	}

	switch c.currentView {
	case RunnerViewForm:
		c.form.SetSize(width, height)
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
			if runner.currentView == RunnerViewForm {
				switch runner.currentPage.Type {
				case types.ListPage:
					runner.currentView = RunnerViewList
					return runner, nil
				case types.DetailPage:
					runner.currentView = RunnerViewDetail
					return runner, nil
				}
			}

			if runner.currentView == RunnerViewLoading {
				return runner, func() tea.Msg {
					return PopPageMsg{}
				}
			}
		}
	case *types.Page:
		runner.SetIsloading(false)
		page := msg
		if page.Title == "" {
			page.Title = "Sunbeam"
		}

		runner.currentPage = page
		switch page.Type {
		case types.DetailPage:
			var detailFunc func() string
			if page.Preview.Text != "" {
				detailFunc = func() string {
					return page.Preview.Text
				}
			} else if page.Preview.Command != "" {
				detailFunc = func() string {
					output, err := utils.RunCommand(page.Preview.Command, strings.NewReader(""), page.Preview.Dir)
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

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				page.Actions = nil
			}

			runner.currentView = RunnerViewDetail
			runner.detail = NewDetail(page.Title, detailFunc, page.Actions)
			runner.detail.Language = page.Preview.Language
			runner.detail.SetSize(runner.width, runner.height)

			return runner, runner.detail.Init()
		case types.FormPage:
			runner.currentView = RunnerViewForm
			var items []FormItem
			for i, input := range page.SubmitAction.Inputs {
				item, err := NewFormItem(input)
				if err != nil {
					return runner, func() tea.Msg {
						return err
					}
				}
				items[i] = item
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				page.SubmitAction = nil
			}

			form, err := NewForm(page.SubmitAction)
			if err != nil {
				return runner, func() tea.Msg {
					return err
				}
			}
			runner.form = form
			runner.form.SetSize(runner.width, runner.height)
			return runner, runner.form.Init()
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
				if !isatty.IsTerminal(os.Stdout.Fd()) {
					listItem.Actions = nil
				}
				listItems[i] = listItem
			}

			cmd := runner.list.SetItems(listItems, selectedId)

			runner.list.SetSize(runner.width, runner.height)
			return runner, tea.Sequence(runner.list.Init(), cmd)
		}

	case types.Action:
		if len(msg.Inputs) > 0 {
			runner.currentView = RunnerViewForm

			form, err := NewForm(&msg)
			if err != nil {
				return runner, func() tea.Msg {
					return err
				}
			}

			runner.form = form
			runner.SetSize(runner.width, runner.height)
			return runner, form.Init()
		}

		cmd := runner.handleAction(msg)
		return runner, cmd
	case error:
		errorView := NewDetail("Error", func() string {
			return fmt.Sprintf("%s", msg)
		}, []types.Action{
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

	if runner.currentView == RunnerViewForm {
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
	if c.err != nil {
		return c.err.View()
	}

	switch c.currentView {
	case RunnerViewForm:
		return c.form.View()
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
