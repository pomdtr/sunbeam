package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/utils"
)

type CommandRunner struct {
	width, height int
	currentView   string

	extension *app.Extension
	command   app.Command

	with map[string]app.Arg

	header Header
	footer Footer

	list   *List
	detail *Detail
	form   *Form
}

func NewCommandRunner(extension *app.Extension, command app.Command, with map[string]app.Arg) *CommandRunner {
	runner := CommandRunner{
		header:      NewHeader(),
		footer:      NewFooter(extension.Title),
		extension:   extension,
		currentView: "loading",
		command:     command,
	}

	// Copy the map to avoid modifying the original
	runner.with = make(map[string]app.Arg)
	for name, arg := range with {
		runner.with[name] = arg
	}

	return &runner
}
func (c *CommandRunner) Init() tea.Cmd {
	return tea.Batch(c.SetIsloading(true), c.Run())
}

type CommandOutput []byte

func (c *CommandRunner) CheckArgs() (map[string]any, []FormItem) {
	args := make(map[string]any)
	formitems := make([]FormItem, 0)
	environ := c.LoadEnv()
	for _, param := range c.command.Params {
		arg, ok := c.with[param.Name]
		if !ok {
			if env, ok := os.LookupEnv(param.Env); ok {
				args[param.Name] = env
				continue
			}

			if env, ok := environ[param.Env]; ok {
				args[param.Name] = env
				continue
			}

			formitems = append(formitems, NewFormItem(param.Name, param.FormItem()))
			continue
		}

		if arg.Value != nil {
			args[param.Name] = fmt.Sprintf("%v", arg.Value)
			continue
		}

		formitems = append(formitems, NewFormItem(param.Name, arg.Input))
	}
	return args, formitems
}

func (c *CommandRunner) LoadEnv() map[string]string {
	dotenvPath := path.Join(c.extension.Root, ".env")
	if _, err := os.Stat(dotenvPath); os.IsNotExist(err) {
		return map[string]string{}
	}

	environ, err := godotenv.Read(dotenvPath)
	if err != nil {
		return map[string]string{}
	}

	return environ
}

func (c *CommandRunner) Run() tea.Cmd {
	// Show form if some parameters are set as input
	args, formItems := c.CheckArgs()
	if len(formItems) > 0 {
		c.currentView = "form"
		c.form = NewForm("args", c.extension.Title, formItems)

		c.form.SetSize(c.width, c.height)
		return c.form.Init()
	}

	payload := app.CmdPayload{
		Args: args,
		Dir:  c.extension.Root,
	}

	if c.currentView == "list" {
		payload.Query = c.list.Query()
	}

	cmd, err := c.command.Cmd(payload)
	if err != nil {
		return NewErrorCmd(err)
	}

	if c.command.OnSuccess == "" {
		return func() tea.Msg {
			return cmd
		}
	}

	return func() tea.Msg {
		output, err := cmd.Output()
		if err != nil {
			exitError, ok := err.(*exec.ExitError)
			if !ok {
				return err
			}

			return fmt.Errorf("%s", exitError.Stderr)
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
	case "form":
		return c.form.SetIsLoading(isLoading)
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
	case "form":
		c.form.SetSize(width, height)
	}
}

func (c *CommandRunner) Reset() {
	c.currentView = "loading"
	c.list = nil
	c.detail = nil
	c.form = nil
}

func (c *CommandRunner) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if c.currentView != "loading" {
				break
			}
			return c, PopCmd
		}
	case CommandOutput:
		c.SetIsloading(false)
		switch c.command.OnSuccess {
		case "push-page":
			var page app.Page
			var v any
			if err := json.Unmarshal(msg, &v); err != nil {
				return c, NewErrorCmd(err)
			}

			if err := app.PageSchema.Validate(v); err != nil {
				return c, NewErrorCmd(err)
			}

			err := json.Unmarshal(msg, &page)
			if err != nil {
				return c, NewErrorCmd(err)
			}

			if page.Title == "" {
				page.Title = c.extension.Title
			}

			switch page.Type {
			case "detail":
				c.currentView = "detail"
				c.detail = NewDetail(page.Title, func() string {
					if page.Detail.Content.Text != "" {
						content, err := utils.HighlightString(page.Detail.Content.Text, page.Detail.Content.Language)
						if err != nil {
							return err.Error()
						}
						return content
					}

					command, ok := c.extension.GetCommand(page.Detail.Content.Command)
					if !ok {
						return fmt.Sprintf("Command %s not found", page.Detail.Content.Command)
					}

					cmd, err := command.Cmd(app.CmdPayload{
						Args: page.Detail.Content.With,
						Dir:  c.extension.Root,
					})
					if err != nil {
						return err.Error()
					}

					output, err := cmd.Output()
					if err != nil {
						return err.Error()
					}

					return string(output)
				})

				actions := make([]Action, len(page.Actions))
				for i, scriptAction := range page.Actions {
					actions[i] = NewAction(scriptAction)
				}
				c.detail.SetActions(actions...)
				c.detail.SetSize(c.width, c.height)

				return c, c.detail.Init()
			case "list":
				c.currentView = "list"

				if c.list == nil {
					c.list = NewList(page.Title)
				} else {
					c.list.SetTitle(page.Title)
				}
				c.list.SetEmptyMessage(page.List.EmptyMessage)

				c.list.defaultActions = make([]Action, len(page.Actions))
				for i, action := range page.Actions {
					c.list.defaultActions[i] = NewAction(action)
				}

				if page.List.ShowPreview {
					c.list.ShowPreview = true
				}
				if page.List.GenerateItems {
					c.list.GenerateItems = true
				}

				listItems := make([]ListItem, len(page.List.Items))
				for i, scriptItem := range page.List.Items {
					scriptItem := scriptItem

					if scriptItem.Id == "" {
						scriptItem.Id = strconv.Itoa(i)
					}
					listItem := ParseScriptItem(scriptItem)
					if scriptItem.Preview != nil {
						listItem.PreviewFunc = func() string {
							if scriptItem.Preview.Command == "" {
								preview, err := utils.HighlightString(scriptItem.Preview.Text, scriptItem.Preview.Language)

								if err != nil {
									return err.Error()
								}
								return preview
							}

							command, ok := c.extension.GetCommand(scriptItem.Preview.Command)
							if !ok {
								return fmt.Sprintf("command %s not found", scriptItem.Preview.Command)
							}

							cmd, err := command.Cmd(app.CmdPayload{
								Args: scriptItem.Preview.With,
								Dir:  c.extension.Root,
							})
							if err != nil {
								return err.Error()
							}

							output, err := cmd.Output()
							if err != nil {
								return err.Error()
							}

							preview, err := utils.HighlightString(string(output), scriptItem.Preview.Language)
							if err != nil {
								return err.Error()
							}

							return preview
						}
					}
					listItems[i] = listItem
				}
				c.list.SetItems(listItems)

				c.list.SetSize(c.width, c.height)
				return c, c.list.Init()
			}
		case "open-url":
			return c, NewOpenUrlCmd(string(msg))
		case "copy-text":
			return c, NewCopyTextCmd(string(msg))
		case "reload-page":
			return c, tea.Sequence(PopCmd, NewReloadPageCmd(nil))
		default:
			return c, Exit
		}

	case SubmitFormMsg:
		switch msg.Id {
		case "args":
			environ := c.LoadEnv()
			for _, param := range c.command.Params {
				value, ok := msg.Values[param.Name]
				if !ok {
					continue
				}

				if param.Env != "" {
					switch value := value.(type) {
					case bool:
						environ[param.Env] = strconv.FormatBool(value)
					case string:
						environ[param.Env] = value
					default:
						return c, NewErrorCmd(fmt.Errorf("invalid value type for param %s", param.Name))
					}

					continue
				}

				arg := c.with[param.Name]
				arg.Value = value
				c.with[param.Name] = arg
			}

			dotenvPath := path.Join(c.extension.Root, ".env")
			if err := godotenv.Write(environ, dotenvPath); err != nil {
				return c, NewErrorCmd(err)
			}

			c.currentView = "loading"

			return c, tea.Sequence(c.SetIsloading(true), c.Run())
		}
	case RunCommandMsg:
		command, ok := c.extension.GetCommand(msg.Command)
		if !ok {
			return c, NewErrorCmd(fmt.Errorf("command not found: %s", msg.Command))
		}
		if msg.OnSuccess != "" {
			command.OnSuccess = msg.OnSuccess
		}

		return c, NewPushCmd(NewCommandRunner(c.extension, command, msg.With))

	case ReloadPageMsg:
		for key, value := range msg.With {
			c.with[key] = value
		}

		return c, tea.Sequence(c.SetIsloading(true), c.Run())
	}

	var cmd tea.Cmd
	var container Page

	switch c.currentView {
	case "list":
		container, cmd = c.list.Update(msg)
		c.list, _ = container.(*List)
	case "detail":
		container, cmd = c.detail.Update(msg)
		c.detail, _ = container.(*Detail)
	case "form":
		container, cmd = c.form.Update(msg)
		c.form, _ = container.(*Form)
	default:
		c.header, cmd = c.header.Update(msg)
	}
	return c, cmd
}

func (c *CommandRunner) View() string {
	switch c.currentView {
	case "list":
		return c.list.View()
	case "detail":
		return c.detail.View()
	case "form":
		return c.form.View()
	case "loading":
		headerView := c.header.View()
		footerView := c.footer.View()
		padding := make([]string, utils.Max(0, c.height-lipgloss.Height(headerView)-lipgloss.Height(footerView)))
		return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), strings.Join(padding, "\n"), c.footer.View())
	default:
		return ""
	}
}
