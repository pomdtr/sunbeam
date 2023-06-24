package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/alessio/shellescape"
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
	case types.PasteAction:
		return func() tea.Msg {
			return ExitMsg{
				Text: action.Text,
			}
		}
	case types.PushAction:
		return func() tea.Msg {
			generator, err := GeneratorFromAction(action)
			if err != nil {
				return err
			}

			return PushPageMsg{
				Page: NewRunner(generator),
			}
		}

	case types.RunAction, types.FetchAction, types.EvalAction:
		return func() tea.Msg {
			output, err := action.Output(runner.ctx)
			if err != nil {
				return err
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
			case types.PasteOnSuccess:
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
			if runner.form != nil && runner.currentPage.Type != types.FormPage {
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
		case types.FormPage:
			form := NewForm(page.Title, func(values map[string]string) tea.Cmd {
				action := *page.SubmitAction
				for key, value := range values {
					action = RenderAction(action, fmt.Sprintf("${input:%s}", key), value)
				}

				return runner.handleAction(action)
			}, page.SubmitAction.Inputs...)

			runner.form = form
			runner.form.SetSize(runner.width, runner.height)

			return runner, runner.form.Init()
		case types.DetailPage:
			detailFunc := func() string {
				if page.Preview == nil {
					return ""
				}
				if page.Preview.Text != "" {
					return page.Preview.Text
				}
				output, err := page.Preview.Command.Output(context.TODO())
				if err != nil {
					return err.Error()
				}
				return string(output)
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

		queryCmd := RenderPageProvider(runner.currentPage.OnQueryChange, "${query}", msg.Query)
		runner.Generator = NewPageProviderGenerator(queryCmd)

		return runner, tea.Sequence(runner.SetIsloading(true), runner.Refresh)
	case types.Action:
		if len(msg.Inputs) > 0 {

			form := NewForm(msg.Title, func(values map[string]string) tea.Cmd {
				return func() tea.Msg {
					submitAction := msg
					for key, value := range values {
						submitAction = RenderAction(submitAction, fmt.Sprintf("${input:%s}", key), value)
					}
					submitAction.Inputs = nil

					return submitAction
				}
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

func RenderRequest(request *types.Request, old, new string) *types.Request {
	rendered := types.Request{}
	rendered.Method = strings.ReplaceAll(request.Method, old, new)
	rendered.Url = strings.ReplaceAll(request.Url, old, url.QueryEscape(new))
	rendered.Body = bytes.ReplaceAll(request.Body, []byte(old), []byte(new))

	for key, value := range request.Headers {
		rendered.Headers[key] = strings.ReplaceAll(value, old, new)
	}

	return &rendered
}

func RenderAction(action types.Action, old, new string) types.Action {
	if action.Command != nil {
		action.Command = RenderCommand(action.Command, old, new)
	}

	if action.Request != nil {
		action.Request = RenderRequest(action.Request, old, new)
	}

	if action.Expression != nil {
		action.Expression = RenderExpression(action.Expression, old, new)
	}

	action.Target = strings.ReplaceAll(action.Target, old, url.QueryEscape(new))
	action.Text = strings.ReplaceAll(action.Text, old, new)
	action.Page = strings.ReplaceAll(action.Page, old, new)
	return action
}

func RenderPreviewProvider(previewProvider *types.PreviewProvider, old, new string) *types.PreviewProvider {
	if previewProvider.Text != "" {
		previewProvider.Text = strings.ReplaceAll(previewProvider.Text, old, new)
	} else if previewProvider.File != "" {
		previewProvider.File = strings.ReplaceAll(previewProvider.File, old, new)
	} else if previewProvider.Command != nil {
		previewProvider.Command = RenderCommand(previewProvider.Command, old, new)
	} else if previewProvider.Request != nil {
		previewProvider.Request = RenderRequest(previewProvider.Request, old, new)
	} else if previewProvider.Expression != nil {
		previewProvider.Expression = RenderExpression(previewProvider.Expression, old, new)
	}
	return previewProvider
}

func RenderPageProvider(pageProvider *types.PageProvider, old, new string) *types.PageProvider {
	if pageProvider.Text != "" {
		pageProvider.Text = strings.ReplaceAll(pageProvider.Text, old, new)
	} else if pageProvider.File != "" {
		pageProvider.File = strings.ReplaceAll(pageProvider.File, old, new)
	} else if pageProvider.Command != nil {
		pageProvider.Command = RenderCommand(pageProvider.Command, old, new)
	} else if pageProvider.Request != nil {
		pageProvider.Request = RenderRequest(pageProvider.Request, old, new)
	} else if pageProvider.Expression != nil {
		pageProvider.Expression = RenderExpression(pageProvider.Expression, old, new)
	}
	return pageProvider
}

func RenderExpression(expression *types.Expression, old, new string) *types.Expression {
	s := string(*expression)
	rendered := types.Expression(strings.ReplaceAll(s, old, new))
	return &rendered
}

func expandPage(page types.Page, base *url.URL) (*types.Page, error) {
	basePath := ""
	if base != nil {
		basePath = base.Path
	}

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

		if action.Command != nil && !path.IsAbs(action.Command.Dir) {
			action.Command.Dir = path.Join(basePath, action.Command.Dir)
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

		if action.Request != nil {
			u, err := expandUri(action.Request.Url)
			if err != nil {
				return nil, err
			}
			action.Request.Url = u
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
			case types.EvalAction:
				action.Title = "Eval"
			case types.ReloadAction:
				action.Title = "Reload"
			case types.FetchAction:
				action.Title = "Fetch"
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

	if page.Preview != nil {
		if page.Preview.Command != nil {
			page.Preview.Command.Dir = basePath
		}
	}

	for i, item := range page.Items {
		if item.Preview != nil {
			if item.Preview.Command != nil {
				item.Preview.Command.Dir = basePath
			}
		}

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
