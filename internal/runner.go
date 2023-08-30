package internal

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/browser"
	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/types"
	"github.com/pomdtr/sunbeam/utils"
)

type CommandRunner struct {
	width, height int
	currentPage   *types.Page

	Generator PageGenerator
	ctx       context.Context
	cancel    context.CancelFunc

	header Header
	footer Footer

	form   *Form
	list   *List
	detail *Detail
	err    *Detail
}

func NewRunner(generator PageGenerator) *CommandRunner {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &CommandRunner{
		header:    NewHeader(),
		footer:    NewFooter("Sunbeam"),
		Generator: generator,
		ctx:       ctx,
		cancel:    cancelFunc,
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
		c.form = nil
		return c.list.Focus()
	case types.DetailPage:
		c.form = nil
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
	switch action.Type {
	case types.ReloadAction:
		runner.form = nil
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

			if action.Exit {
				return ExitMsg{}
			}

			return nil
		}
	case types.CopyAction:
		return func() tea.Msg {
			err := clipboard.WriteAll(action.Text)
			if err != nil {
				return err
			}

			if action.Exit {
				return ExitMsg{}
			}

			return nil
		}
	case types.PipeAction:
		return func() tea.Msg {
			return ExitMsg{
				Text: action.Text,
			}
		}
	case types.PushAction:
		return func() tea.Msg {
			generator := NewCommandGenerator(action.Command)

			return PushPageMsg{
				Page: NewRunner(generator),
			}
		}

	case types.RunAction:
		return func() tea.Msg {
			if action.Command != nil && action.OnSuccess == "" {
				cmd, err := action.Command.Cmd(context.TODO())
				if err != nil {
					return err
				}

				return ExitMsg{
					Cmd: cmd,
				}
			}

			output, err := action.Output(runner.ctx)
			if err != nil {
				return err
			}

			if action.OnSuccess == "" {
				return ExitMsg{
					Text: string(output),
				}
			}

			switch action.OnSuccess {
			case types.CopyOnSuccess:
				if err := clipboard.WriteAll(string(output)); err != nil {
					return err
				}

				if action.Exit {
					return ExitMsg{}
				}

				return nil
			case types.PipeOnSuccess:
				return ExitMsg{
					Text: string(output),
				}
			case types.OpenOnSuccess:
				if err := browser.OpenURL(string(output)); err != nil {
					return err
				}

				if action.Exit {
					return ExitMsg{}
				}

				return nil
			case types.ReloadOnSuccess:
				return types.Action{
					Type: types.ReloadAction,
				}
			default:
				return fmt.Errorf("unknown on_success action: %s", action.OnSuccess)
			}
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

func (runner CommandRunner) IsLoading() bool {
	return runner.currentPage == nil
}

func (runner *CommandRunner) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// if form is shown over a page, close it
			if runner.form != nil {
				runner.form = nil
				return runner, nil
			}

			if runner.IsLoading() {
				return runner, func() tea.Msg {
					runner.cancel()
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
			runner.detail = NewDetail(page.Title, page.Text, page.Actions)
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
				listItems[i] = listItem
			}

			runner.list.SetItems(listItems, selectedId)

			runner.list.SetSize(runner.width, runner.height)
			return runner, runner.list.Init()
		}

	case types.Action:
		// if len(msg.Inputs) > 0 {
		// 	form := NewForm(msg.Title, func(values map[string]string) tea.Cmd {
		// 		submitAction := msg
		// 		for key, value := range values {
		// 			submitAction = RenderAction(submitAction, fmt.Sprintf("{{input:%s}}", key), value)
		// 		}

		// 		return runner.handleAction(submitAction)
		// 	}, msg.Inputs...)

		// 	runner.form = form
		// 	runner.SetSize(runner.width, runner.height)
		// 	return runner, form.Init()
		// }

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

		errorView := NewDetail("Error", content, []types.Action{
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

func expandPage(page types.Page, base *url.URL) (*types.Page, error) {
	expandUri := func(target string) (string, error) {
		if base == nil {
			return target, nil
		}

		targetUrl, err := url.Parse(target)
		if err != nil {
			return "", err
		}

		switch targetUrl.Scheme {
		case "http", "https":
			return target, nil
		case "file":
			return target, nil
		case "":
			res := &url.URL{Scheme: base.Scheme, Host: base.Host, Path: path.Join(base.Path, targetUrl.Path)}
			return res.String(), nil
		default:
			return "", fmt.Errorf("unsupported scheme: %s", targetUrl.Scheme)
		}
	}

	expandAction := func(action types.Action) (*types.Action, error) {
		if action.Type == types.RunAction && base != nil && base.Scheme != "file" && base.Scheme != "" {
			return nil, fmt.Errorf("run action is not supported for remote pages")
		}

		if action.Page != "" {
			p, err := expandUri(action.Page)
			if err != nil {
				return nil, err
			}
			action.Page = p
		}

		if action.Target != "" {
			t, err := expandUri(action.Target)
			if err != nil {
				return nil, err
			}

			action.Target = t
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
			case types.ReloadAction:
				action.Title = "Reload"
			}
		}

		return &action, nil
	}

	for i, action := range page.Actions {
		a, err := expandAction(action)
		if err != nil {
			return nil, err
		}

		page.Actions[i] = *a
	}

	for i, item := range page.Items {
		for j, action := range item.Actions {
			action, err := expandAction(action)
			if err != nil {
				return nil, err
			}

			item.Actions[j] = *action
		}

		page.Items[i] = item
	}

	return &page, nil
}
