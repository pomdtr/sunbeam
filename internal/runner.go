package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
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
)

type PageGenerator func() (*types.Page, error)

func NewFileGenerator(name string) PageGenerator {
	return func() (*types.Page, error) {
		b, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}

		if err := schemas.Validate(b); err != nil {
			return nil, err
		}

		var page types.Page
		if err := json.Unmarshal(b, &page); err != nil {
			return nil, err
		}

		page = expandPage(page, filepath.Dir(name))
		return &page, nil
	}
}

func NewStaticGenerator(reader io.Reader) PageGenerator {
	var pageRef *types.Page
	return func() (*types.Page, error) {
		if pageRef != nil {
			return pageRef, nil
		}

		b, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		if err := schemas.Validate(b); err != nil {
			return nil, err
		}

		var page types.Page
		if err := json.Unmarshal(b, &page); err != nil {
			return nil, err
		}

		page = expandPage(page, "")
		pageRef = &page
		return &page, nil
	}
}

func NewCommandGenerator(command *types.Command) PageGenerator {
	return func() (*types.Page, error) {
		output, err := command.Output()
		if err != nil {
			return nil, err
		}

		if err := schemas.Validate(output); err != nil {
			return nil, err
		}

		var page types.Page
		if err := json.Unmarshal(output, &page); err != nil {
			return nil, err
		}

		page = expandPage(page, "")
		return &page, nil
	}
}

type CommandRunner struct {
	width, height int
	currentPage   *types.Page

	Generator PageGenerator

	header Header
	footer Footer

	form   *Form
	list   *List
	detail *Detail
	err    *Detail
}

func NewRunner(generator PageGenerator) *CommandRunner {
	return &CommandRunner{
		header:    NewHeader(),
		footer:    NewFooter("Sunbeam"),
		Generator: generator,
	}

}
func (c *CommandRunner) Init() tea.Cmd {
	return tea.Batch(c.SetIsloading(true), c.Refresh)
}

func (c *CommandRunner) Focus() tea.Cmd {
	if c.currentPage == nil {
		return nil
	}
	switch c.currentPage.Type {
	case types.ListPage:
		return c.list.Focus()
	case types.DetailPage:
		return c.detail.Focus()
	}

	return nil
}

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
		action = RenderAction(action, fmt.Sprintf("${env:%s}", pair[0]), pair[1])
	}

	switch action.Type {
	case types.ReloadAction:
		if action.Command != nil {
			runner.Generator = NewCommandGenerator(action.Command)
		}

		return tea.Sequence(runner.SetIsloading(true), runner.Refresh)
	case types.OpenAction:
		return func() tea.Msg {
			if err := browser.OpenURL(action.Target); err != nil {
				return func() tea.Msg {
					return err
				}
			}

			return ExitMsg{}
		}
	case types.CopyAction:
		return func() tea.Msg {
			err := clipboard.WriteAll(action.Text)
			if err != nil {
				return err
			}

			return ExitMsg{}
		}

	case types.PushAction:
		return func() tea.Msg {
			if action.Page == nil {
				return errors.New("page is nil")
			}

			if action.Page.Command != nil {
				return PushPageMsg{
					Page: NewRunner(NewCommandGenerator(action.Page.Command)),
				}
			}

			return PushPageMsg{
				Page: NewRunner(NewFileGenerator(action.Page.Text)),
			}
		}

	case types.RunAction:
		return func() tea.Msg {
			if !action.ReloadOnSuccess {
				return ExitMsg{
					Cmd: action.Command.Cmd(),
				}
			}

			if err := action.Command.Run(); err != nil {
				return err
			}

			return types.NewReloadAction()
		}
	default:
		return func() tea.Msg {
			return fmt.Errorf("unknown action type: %s", action.Type)
		}
	}
}

func (c *CommandRunner) SetIsloading(isLoading bool) tea.Cmd {
	if c.currentPage == nil {
		return c.header.SetIsLoading(isLoading)
	}

	switch c.currentPage.Type {
	case types.ListPage:
		return c.list.SetIsLoading(isLoading)
	case types.DetailPage:
		return c.detail.SetIsLoading(isLoading)
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

			if runner.currentPage == nil {
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
			detailFunc := func() string {
				if page.Preview == nil {
					return ""
				}
				if page.Preview.Text != "" {
					return page.Preview.Text
				}
				output, err := page.Preview.Command.Output()
				if err != nil {
					return err.Error()
				}
				return string(output)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				page.Actions = nil
			}

			runner.detail = NewDetail(page.Title, detailFunc, page.Actions)
			runner.detail.SetSize(runner.width, runner.height)

			return runner, runner.detail.Init()
		case types.ListPage:

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
			return runner, tea.Batch(runner.list.Init(), cmd)
		}

	case QueryChangeMsg:
		if runner.currentPage == nil || runner.currentPage.Type != types.ListPage {
			return runner, nil
		}

		queryCmd := RenderCommand(runner.currentPage.OnQueryChange, "${query}", msg.Query)
		runner.Generator = NewCommandGenerator(queryCmd)

		return runner, tea.Sequence(runner.SetIsloading(true), runner.Refresh)
	case types.Action:
		if len(msg.Inputs) > 0 {

			form := NewForm(msg.Title, func(values map[string]string) tea.Msg {
				submitAction := msg
				for key, value := range values {
					submitAction = RenderAction(submitAction, fmt.Sprintf("${input:%s}", key), value)
				}
				submitAction.Inputs = nil

				return submitAction
			}, msg.Inputs...)

			runner.form = form
			runner.SetSize(runner.width, runner.height)
			return runner, form.Init()
		}

		runner.form = nil
		cmd := runner.handleAction(msg)
		return runner, cmd
	case error:
		var ve schemas.ValidationError
		var content string
		if errors.As(msg, &ve) {
			content = ve.Message()
		} else {
			content = msg.Error()
		}

		errorView := NewDetail("Error", func() string {
			return content
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

	if runner.err != nil {
		container, cmd = runner.err.Update(msg)
		runner.err, _ = container.(*Detail)
		return runner, cmd
	}

	if runner.form != nil {
		container, cmd = runner.form.Update(msg)
		runner.form, _ = container.(*Form)
		return runner, cmd
	}

	if runner.currentPage == nil {
		runner.header, cmd = runner.header.Update(msg)
		return runner, cmd
	}

	switch runner.currentPage.Type {
	case types.ListPage:
		container, cmd = runner.list.Update(msg)
		runner.list, _ = container.(*List)
	case types.DetailPage:
		container, cmd = runner.detail.Update(msg)
		runner.detail, _ = container.(*Detail)
	}

	return runner, cmd
}

func (c *CommandRunner) View() string {
	if c.err != nil {
		return c.err.View()
	}

	if c.form != nil {
		return c.form.View()
	}

	if c.currentPage == nil {
		headerView := c.header.View()
		footerView := c.footer.View()
		padding := make([]string, utils.Max(0, c.height-lipgloss.Height(headerView)-lipgloss.Height(footerView)))
		return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), strings.Join(padding, "\n"), c.footer.View())
	}

	switch c.currentPage.Type {
	case types.ListPage:
		return c.list.View()
	case types.DetailPage:
		return c.detail.View()
	default:
		return ""
	}
}

func RenderCommand(command *types.Command, old, new string) *types.Command {
	rendered := types.Command{}
	rendered.Name = strings.ReplaceAll(command.Name, old, new)
	for _, arg := range command.Args {
		rendered.Args = append(rendered.Args, strings.ReplaceAll(arg, old, shellescape.Quote(new)))
	}
	rendered.Input = strings.ReplaceAll(command.Input, old, new)
	rendered.Dir = strings.ReplaceAll(command.Dir, old, new)

	return &rendered
}

func RenderAction(action types.Action, old, new string) types.Action {
	if action.Command != nil {
		action.Command = RenderCommand(action.Command, old, new)
	}

	action.Target = strings.ReplaceAll(action.Target, old, url.QueryEscape(new))
	action.Text = strings.ReplaceAll(action.Text, old, new)
	if action.Page != nil {
		if action.Page.Command != nil {
			action.Page.Command = RenderCommand(action.Page.Command, old, new)
		} else {
			action.Page.Text = strings.ReplaceAll(action.Page.Text, old, new)
		}
	}
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

		if filepath.IsAbs(targetUrl.Path) {
			return target
		}

		return filepath.Join(dir, targetUrl.Path)
	}

	expandAction := func(action types.Action) types.Action {
		if action.Command != nil && !filepath.IsAbs(action.Command.Dir) {
			action.Command.Dir = filepath.Join(dir, action.Command.Dir)
		}

		if action.Page != nil && action.Page.Text != "" && !filepath.IsAbs(action.Page.Text) {
			action.Page.Text = filepath.Join(dir, action.Page.Text)
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
