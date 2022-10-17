package pages

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/bubbles"
	"github.com/pomdtr/sunbeam/scripts"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/skratchdot/open-golang/open"
)

type CommandContainer struct {
	width   int
	height  int
	command scripts.Command
	input   scripts.CommandInput
	// spinner spinner.Model
	embed Page
}

func NewCommandContainer(command scripts.Command, input scripts.CommandInput) *CommandContainer {
	// s := spinner.New()
	// s.Spinner = spinner.Line
	// s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return &CommandContainer{command: command, input: input}
}

func (c *CommandContainer) headerView() string {
	line := strings.Repeat("─", c.width)
	return fmt.Sprintf("\n%s", line)
}

func (c *CommandContainer) SetSize(width, height int) {
	c.width = width
	c.height = height
	if c.embed != nil {
		c.embed.SetSize(width, height)
	}
}

func RunCmd(command scripts.Command, input scripts.CommandInput) tea.Cmd {
	return func() tea.Msg {
		res, err := command.Run(input)
		if err != nil {
			return err
		}
		return res
	}
}

func (c *CommandContainer) Init() tea.Cmd {
	return tea.Batch(RunCmd(c.command, c.input))
}

func (c *CommandContainer) footerView() string {
	return bubbles.SunbeamFooter(c.width, c.command.Title())
}

func (c *CommandContainer) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			if c.embed == nil {
				return c, PopCmd
			}
		}
	case scripts.ScriptResponse:
		switch msg.Type {
		case "list":
			list := msg.List
			if list.Title == "" {
				list.Title = c.command.Title()
			}
			c.embed = NewListPage(msg.List)
			c.embed.SetSize(c.width, c.height)
			return c, c.embed.Init()
		case "detail":
			detail := msg.Detail
			if detail.Title == "" {
				detail.Title = c.command.Title()
			}
			c.embed = NewDetailContainer(msg.Detail, c.RunAction)
			c.embed.SetSize(c.width, c.height)
			return c, c.embed.Init()
		case "form":
			form := msg.Form
			if form.Title == "" {
				form.Title = c.command.Title()
			}
			submitAction := func(values map[string]string) tea.Cmd {
				switch form.Method {
				case "args":
					return RunCmd(c.command, c.input)
				case "env":
					return RunCmd(c.command, c.input)
				}
				return utils.NewErrorCmd("unknown form method: %s", msg.Form.Method)
			}
			c.embed = NewFormContainer(msg.Form, submitAction)
			c.embed.SetSize(c.width, c.height)
			return c, c.embed.Init()
		case "action":
			cmd = c.RunAction(*msg.Action)
			return c, cmd
		}
	case scripts.ScriptAction:
		return c, c.RunAction(msg)
		// case spinner.TickMsg:
		// 	if c.embed != nil {
		// 		return c, nil
		// 	}
		// 	var cmd tea.Cmd
		// 	c.spinner, cmd = c.spinner.Update(msg)
		// 	return c, cmd
	}

	if c.embed != nil {
		c.embed, cmd = c.embed.Update(msg)
	}

	return c, cmd
}

func (container *CommandContainer) View() string {
	if container.embed != nil {
		return container.embed.View()
	}

	var loadingIndicator string
	// spinner := lipgloss.NewStyle().Padding(0, 2).Render(container.spinner.View())
	label := lipgloss.NewStyle().Render("Loading...")
	loadingIndicator = lipgloss.JoinHorizontal(lipgloss.Center, label)

	newLines := strings.Repeat("\n", utils.Max(0, container.height-lipgloss.Height(loadingIndicator)-lipgloss.Height(container.footerView())-lipgloss.Height(container.headerView())-1))

	return lipgloss.JoinVertical(lipgloss.Left, container.headerView(), loadingIndicator, newLines, container.footerView())
}

func (c CommandContainer) RunAction(action scripts.ScriptAction) tea.Cmd {
	switch action.Type {
	case "push":
		scriptPath := path.Join(scripts.CommandDir, action.Path)
		command, err := scripts.Parse(scriptPath)
		if err != nil {
			log.Fatal(err)
		}

		return NewPushCmd(NewCommandContainer(command, scripts.CommandInput{
			Arguments: action.Args,
		}))
	case "exec":
		var cmd *exec.Cmd
		if len(action.Command) == 1 {
			cmd = exec.Command(action.Command[0])
		} else {
			cmd = exec.Command(action.Command[0], action.Command[1:]...)
		}
		err := cmd.Run()
		if err != nil {
			return utils.SendMsg(
				scripts.ScriptResponse{
					Type: "detail",
					Detail: &scripts.DetailResponse{
						Format: "text",
						Text:   err.Error(),
					},
				},
			)
		}
		return tea.Quit
	case "open":
		err := open.Run(action.Path)
		if err != nil {
			return utils.NewErrorCmd("failed to open file: %s", err)
		}
		return tea.Quit
	case "open-url":
		err := open.Run(action.Url)
		if err != nil {
			return utils.NewErrorCmd("failed to open url: %s", err)
		}
		return tea.Quit
	case "copy":
		err := clipboard.WriteAll(action.Content)
		if err != nil {
			return utils.NewErrorCmd("failed to access clipboard: %s", err)
		}
		return tea.Quit
	default:
		log.Printf("Unknown action type: %s", action.Type)
		return tea.Quit
	}

}
