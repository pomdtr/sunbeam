package tui

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/scripts"
	"github.com/pomdtr/sunbeam/utils"
)

type CommandRunner struct {
	width, height int
	currentView   string

	command string
	args    []string

	header Header
	footer Footer

	list   *List
	detail *Detail
}

func NewCommandRunner(command string, args ...string) *CommandRunner {
	runner := CommandRunner{
		header:      NewHeader(),
		footer:      NewFooter("Sunbeam"),
		currentView: "loading",
		command:     command,
		args:        args,
	}

	return &runner
}
func (c *CommandRunner) Init() tea.Cmd {
	return tea.Batch(c.SetIsloading(true), c.Run())
}

type CommandOutput []byte

func (c *CommandRunner) Run() tea.Cmd {
	cmd := exec.Command(c.command, c.args...)

	if c.currentView == "list" {
		cmd.Stdin = strings.NewReader(c.list.Query())
	}

	return func() tea.Msg {
		output, err := cmd.Output()
		if err != nil {
			exitError, ok := err.(*exec.ExitError)
			if !ok {
				return err
			}

			return fmt.Errorf("error running command %s: %s", cmd.String(), exitError.Stderr)
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
		var res scripts.Response
		var v any
		if err := json.Unmarshal(msg, &v); err != nil {
			return c, NewErrorCmd(err)
		}

		if err := scripts.Schema.Validate(v); err != nil {
			return c, NewErrorCmd(err)
		}

		err := json.Unmarshal(msg, &res)
		if err != nil {
			return c, NewErrorCmd(err)
		}

		if res.Title == "" {
			res.Title = "Sunbeam"
		}

		switch res.Type {
		case "detail":
			c.currentView = "detail"
			c.detail = NewDetail(res.Title, func() string {
				if res.Detail.Content.Text != "" {
					return res.Detail.Content.Text
				}

				cmd := exec.Command(res.Detail.Content.Command, res.Detail.Content.Args...)

				output, err := cmd.Output()
				if err != nil {
					return err.Error()
				}

				return string(output)
			})

			actions := make([]Action, len(res.Actions))
			for i, scriptAction := range res.Actions {
				actions[i] = NewAction(scriptAction)
			}
			c.detail.SetActions(actions...)
			c.detail.SetSize(c.width, c.height)

			return c, c.detail.Init()
		case "list":
			c.currentView = "list"

			if c.list == nil {
				c.list = NewList(res.Title)
			} else {
				c.list.SetTitle(res.Title)
			}

			if res.List.ShowPreview {
				c.list.ShowPreview = true
			}
			if res.List.GenerateItems {
				c.list.GenerateItems = true
			}

			listItems := make([]ListItem, len(res.List.Items))
			for i, scriptItem := range res.List.Items {
				scriptItem := scriptItem

				if scriptItem.Id == "" {
					scriptItem.Id = strconv.Itoa(i)
				}
				listItem := ParseScriptItem(scriptItem)
				if scriptItem.Preview != nil {
					listItem.PreviewFunc = func() string {
						if scriptItem.Preview.Text != "" {
							return scriptItem.Preview.Text
						}

						cmd := exec.Command(scriptItem.Preview.Command, scriptItem.Preview.Args...)
						output, err := cmd.Output()
						if err != nil {
							return err.Error()
						}
						return string(output)
					}
				}
				listItems[i] = listItem
			}
			c.list.SetItems(listItems)

			c.list.SetSize(c.width, c.height)
			return c, c.list.Init()
		}

	case ReloadPageMsg:
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
	case "loading":
		headerView := c.header.View()
		footerView := c.footer.View()
		padding := make([]string, utils.Max(0, c.height-lipgloss.Height(headerView)-lipgloss.Height(footerView)))
		return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), strings.Join(padding, "\n"), c.footer.View())
	default:
		return ""
	}
}
