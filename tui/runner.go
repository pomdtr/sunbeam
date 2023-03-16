package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/shlex"
	"github.com/pkg/browser"
	"github.com/pomdtr/sunbeam/types"
	"github.com/pomdtr/sunbeam/utils"
	"mvdan.cc/sh/v3/shell"
)

type CommandRunner struct {
	width, height int
	currentView   RunnerView

	Generator  PageGenerator
	Validator  PageValidator
	workingDir string

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
	RunnerViewError
	RunnerViewForm
)

type PageValidator func([]byte) error

func NewRunner(generator PageGenerator, validator PageValidator, workingDir string) *CommandRunner {
	return &CommandRunner{
		header:      NewHeader(),
		footer:      NewFooter("Sunbeam"),
		currentView: RunnerViewLoading,
		Generator:   generator,
		Validator:   validator,
		workingDir:  workingDir,
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
	switch action.Type {
	case types.ReloadAction:
		return tea.Sequence(runner.SetIsloading(true), runner.Refresh)
	case types.EditAction:
		if strings.HasPrefix(action.Path, "~") {
			home, _ := os.UserHomeDir()
			action.Path = path.Join(home, action.Path[1:])
		}
		return func() tea.Msg {
			return types.Action{
				Type:      types.RunAction,
				OnSuccess: types.ExitOnSuccess,
				Command:   fmt.Sprintf("${EDITOR:-vi} %s", shellescape.Quote(action.Path)),
			}
		}
	case types.OpenAction:
		var target string
		if action.Url != "" {
			target = action.Url
		} else {
			if strings.HasPrefix(action.Path, "~") {
				home, _ := os.UserHomeDir()
				target = path.Join(home, action.Path[1:])
			} else if !path.IsAbs(action.Path) {
				target = path.Join(runner.workingDir, action.Path)
			} else {
				target = action.Path
			}
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
	case types.ReadAction:
		var page string
		if path.IsAbs(action.Path) {
			page = action.Path
		} else if strings.HasPrefix(action.Path, "~") {
			home, _ := os.UserHomeDir()
			page = path.Join(home, action.Path[1:])
		} else {
			page = path.Join(runner.workingDir, action.Path)
		}

		runner := NewRunner(NewFileGenerator(
			page,
		), runner.Validator, path.Dir(page))
		return func() tea.Msg {
			return PushPageMsg{runner: runner}
		}
	case types.RunAction:
		args, err := shlex.Split(action.Command)
		if err != nil {
			return func() tea.Msg {
				return fmt.Errorf("failed to parse command: %s", err)
			}
		}

		name := args[0]
		var extraArgs []string
		if len(args) > 1 {
			extraArgs = args[1:]
		}

		switch action.OnSuccess {
		case types.PushOnSuccess:
			workingDir := runner.workingDir
			generator := NewCommandGenerator(name, extraArgs, workingDir)
			runner := NewRunner(generator, runner.Validator, workingDir)
			return func() tea.Msg {
				return PushPageMsg{runner: runner}
			}
		case types.ReloadOnSuccess:
			return func() tea.Msg {
				command := exec.Command(name, extraArgs...)
				command.Dir = runner.workingDir
				err := command.Run()
				if err != nil {
					if err, ok := err.(*exec.ExitError); ok {
						return fmt.Errorf("command exit with code %d: %s", err.ExitCode(), err.Stderr)
					}
					return err
				}

				return types.Action{
					Type: types.ReloadAction,
				}
			}
		case types.ReplaceOnSuccess:
			runner.Generator = NewCommandGenerator(name, extraArgs, runner.workingDir)
			return runner.Refresh

		case types.ExitOnSuccess:
			command := exec.Command(name, extraArgs...)
			command.Dir = runner.workingDir

			return func() tea.Msg {
				return command
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

	switch c.currentView {
	case RunnerViewList:
		c.list.SetSize(width, height)
	case RunnerViewDetail:
		c.detail.SetSize(width, height)
	case RunnerViewForm:
		c.form.SetSize(width, height)
	case RunnerViewError:
		c.err.SetSize(width, height)
	}
}

func (runner *CommandRunner) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if runner.currentView != RunnerViewLoading {
				break
			}
			return runner, func() tea.Msg {
				return PopPageMsg{}
			}
		}
	case CommandOutput:
		runner.SetIsloading(false)
		if err := runner.Validator(msg); err != nil {
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
			runner.currentView = RunnerViewDetail

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

	case types.Action:
		if len(msg.Inputs) > 0 {
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

			form := NewForm(formItems, func(values map[string]string) tea.Cmd {
				command := msg.Command

				for key, value := range values {
					command = strings.ReplaceAll(command, fmt.Sprintf("${input:%s}", key), value)
				}

				return func() tea.Msg {
					return types.Action{
						Type:      types.RunAction,
						Command:   command,
						OnSuccess: msg.OnSuccess,
					}
				}
			})

			runner.currentView = RunnerViewForm
			runner.form = form
			runner.SetSize(runner.width, runner.height)
			return runner, form.Init()
		}

		cmd := runner.handleAction(msg)
		return runner, cmd
	case error:
		runner.currentView = RunnerViewError
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
	case RunnerViewList:
		container, cmd = runner.list.Update(msg)
		runner.list, _ = container.(*List)
	case RunnerViewDetail:
		container, cmd = runner.detail.Update(msg)
		runner.detail, _ = container.(*Detail)
	case RunnerViewForm:
		container, cmd = runner.form.Update(msg)
		runner.form, _ = container.(*Form)
	case RunnerViewError:
		container, cmd = runner.err.Update(msg)
		runner.err, _ = container.(*Detail)
	default:
		runner.header, cmd = runner.header.Update(msg)
	}
	return runner, cmd
}

func (c *CommandRunner) View() string {
	switch c.currentView {
	case RunnerViewList:
		return c.list.View()
	case RunnerViewForm:
		return c.form.View()
	case RunnerViewDetail:
		return c.detail.View()
	case RunnerViewError:
		return c.err.View()
	case RunnerViewLoading:
		headerView := c.header.View()
		footerView := c.footer.View()
		padding := make([]string, utils.Max(0, c.height-lipgloss.Height(headerView)-lipgloss.Height(footerView)))
		return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), strings.Join(padding, "\n"), c.footer.View())
	default:
		return ""
	}
}
