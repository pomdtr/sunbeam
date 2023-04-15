package internal

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/pkg/browser"
	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/types"
	"github.com/pomdtr/sunbeam/utils"
	"gopkg.in/yaml.v3"
)

type PageGenerator func() ([]byte, error)

func NewFileGenerator(name string) PageGenerator {
	return func() ([]byte, error) {
		extension := filepath.Ext(name)
		bytes, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}

		var page types.Page
		if extension == ".yaml" || extension == ".yml" {
			if err := yaml.Unmarshal(bytes, &page); err != nil {
				return nil, err
			}
		} else {
			if err := json.Unmarshal(bytes, &page); err != nil {
				return nil, err
			}
		}

		page = expandPage(page, filepath.Dir(name))
		return json.Marshal(page)
	}
}

func NewCommandGenerator(command *types.Command) PageGenerator {
	return func() ([]byte, error) {
		output, err := command.Output()
		if err != nil {
			return nil, err
		}

		var page types.Page
		if err := json.Unmarshal(output, &page); err != nil {
			return nil, err
		}

		page = expandPage(page, command.Dir)
		return json.Marshal(page)
	}
}

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
	b, err := runner.Generator()
	if err != nil {
		return err
	}

	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	if err := schemas.Validate(v); err != nil {
		return err
	}

	var page types.Page
	if err := json.Unmarshal(b, &page); err != nil {
		return err
	}

	return &page
}

func (runner *CommandRunner) handleAction(action types.Action) tea.Cmd {
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		action = RenderAction(action, fmt.Sprintf("${env:%s}", pair[0]), pair[1])
	}

	switch action.Type {
	case types.ReloadAction:
		return tea.Sequence(runner.SetIsloading(true), runner.Refresh)
	case types.OpenAction:
		return func() tea.Msg {
			if err := browser.OpenURL(action.Target); err != nil {
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

	case types.PushAction:
		return func() tea.Msg {
			return PushPageMsg{
				runner: NewRunner(NewFileGenerator(action.Page)),
			}
		}

	case types.RunAction:
		switch action.OnSuccess {
		case types.OpenOnSuccess:
			return func() tea.Msg {
				output, err := action.Command.Output()
				if err != nil {
					return err
				}
				if err := browser.OpenURL(string(output)); err != nil {
					return fmt.Errorf("failed to open url: %s", err)
				}
				return tea.Quit()
			}
		case types.CopyOnSuccess:
			return func() tea.Msg {
				output, err := action.Command.Output()
				if err != nil {
					return err
				}

				if err := clipboard.WriteAll(string(output)); err != nil {
					return fmt.Errorf("failed to copy text to clipboard: %s", err)
				}
				return tea.Quit()
			}

		case types.PushOnSuccess:
			return func() tea.Msg {
				return PushPageMsg{
					runner: NewRunner(NewCommandGenerator(action.Command)),
				}
			}

		case types.ReloadOnSuccess:
			return func() tea.Msg {
				err := action.Command.Run()
				if err != nil {
					return err
				}
				return types.ReloadAction
			}
		case types.ExitOnSuccess:
			return func() tea.Msg {
				err := action.Command.Run()
				if err != nil {
					return err
				}
				return tea.Quit()
			}
		default:
			return func() tea.Msg {
				err := action.Command.Run()
				if err != nil {
					return err
				}
				return tea.Quit()
			}
		}

	default:
		return func() tea.Msg {
			return fmt.Errorf("unknown action type: %s", action.Type)
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
			} else {
				detailFunc = func() string {
					output, err := page.Preview.Command.Output()
					if err != nil {
						return err.Error()
					}
					return string(output)
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
			items := make([]FormItem, len(page.SubmitAction.Inputs))
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

func RenderAction(action types.Action, old, new string) types.Action {
	if action.Command != nil {
		for i, arg := range action.Command.Args {
			action.Command.Args[i] = strings.ReplaceAll(arg, old, shellescape.Quote(new))
		}
		action.Command.Input = strings.ReplaceAll(action.Command.Input, old, new)
		action.Command.Dir = strings.ReplaceAll(action.Command.Dir, old, new)
	}

	action.Target = strings.ReplaceAll(action.Target, old, url.QueryEscape(new))
	action.Text = strings.ReplaceAll(action.Text, old, new)
	action.Page = strings.ReplaceAll(action.Page, old, new)
	return action
}

func expandPage(page types.Page, dir string) types.Page {
	expandUrl := func(target string) string {
		targetUrl, err := url.Parse(target)
		if err != nil {
			return target
		}

		if targetUrl.Scheme != "" && targetUrl.Scheme != "file" {
			return target
		}

		if path.IsAbs(targetUrl.Path) {
			return target
		}

		return path.Join(dir, targetUrl.Path)
	}

	expandAction := func(action types.Action) types.Action {
		if action.Command != nil && !path.IsAbs(action.Command.Dir) {
			action.Command.Dir = path.Join(dir, action.Command.Dir)
		}

		if action.Page != "" && !path.IsAbs(action.Page) {
			action.Page = path.Join(dir, action.Page)
		}

		if action.Target != "" {
			action.Target = expandUrl(action.Target)
		}

		if action.Title == "" {
			switch action.Type {
			case types.OpenAction:
				action.Title = "Open"
			case types.CopyAction:
				action.Title = "Copy"
			case types.RunAction:
				action.Title = "Run"
			case types.PushAction:
				action.Title = "Push"
			case types.ExitAction:
				action.Title = "Exit"
			case types.ReloadAction:
				action.Title = "Reload"
			}
		}

		return action
	}

	for i, action := range page.Actions {
		page.Actions[i] = expandAction(action)
	}

	if page.Preview != nil {
		if page.Preview.Command != nil {
			page.Preview.Command.Dir = dir
		}
	}

	for i, item := range page.Items {
		if item.Preview != nil {
			if item.Preview.Command != nil {
				item.Preview.Command.Dir = dir
			}
		}

		for j, action := range item.Actions {

			item.Actions[j] = expandAction(action)
		}

		page.Items[i] = item
	}

	return page
}
