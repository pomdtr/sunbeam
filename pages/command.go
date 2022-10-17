package pages

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
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
	spinner spinner.Model
	embed   Page
}

func NewCommandContainer(command scripts.Command) *CommandContainer {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return &CommandContainer{command: command, spinner: s}
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

func (c *CommandContainer) Init() tea.Cmd {
	return tea.Batch(c.spinner.Tick, c.fetchItems(c.command))
}

func (c CommandContainer) fetchItems(command scripts.Command) tea.Cmd {
	return func() tea.Msg {
		res, err := command.Run()
		if err != nil {
			return err
		}
		return res
	}
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
			c.embed = NewListContainer(msg.List)
			c.embed.SetSize(c.width, c.height)
		case "detail":
			detail := msg.Detail
			if detail.Title == "" {
				detail.Title = c.command.Title()
			}
			c.embed = NewDetailContainer(msg.Detail, c.RunAction)
			c.embed.SetSize(c.width, c.height)
		case "form":
			form := msg.Form
			if form.Title == "" {
				form.Title = c.command.Title()
			}
			submitAction := func(values map[string]string) tea.Cmd {
				switch form.Method {
				case "args":
					for _, arg := range c.command.Metadatas.Arguments {
						c.command.Arguments = append(c.command.Arguments, values[arg.Placeholder])
					}
					return c.fetchItems(c.command)
				case "env":
					c.command.Environment = values
					return c.fetchItems(c.command)
				}
				return utils.NewErrorCmd("unknown form method: %s", msg.Form.Method)
			}
			c.embed = NewFormContainer(msg.Form, submitAction)
			c.embed.SetSize(c.width, c.height)
		case "action":
			cmd = c.RunAction(*msg.Action)
			return c, cmd
		}
	case scripts.ScriptAction:
		return c, c.RunAction(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		c.spinner, cmd = c.spinner.Update(msg)
		return c, cmd
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
	spinner := lipgloss.NewStyle().Padding(0, 2).Render(container.spinner.View())
	label := lipgloss.NewStyle().Render("Loading...")
	loadingIndicator = lipgloss.JoinHorizontal(lipgloss.Center, spinner, label)

	newLines := strings.Repeat("\n", utils.Max(0, container.height-lipgloss.Height(loadingIndicator)-lipgloss.Height(container.footerView())-lipgloss.Height(container.headerView())-1))

	return lipgloss.JoinVertical(lipgloss.Left, container.headerView(), loadingIndicator, newLines, container.footerView())
}

func (c CommandContainer) RunAction(action scripts.ScriptAction) tea.Cmd {
	switch action.Type {
	case "push":
		commandDir := path.Dir(c.command.Url.Path)
		scriptPath := path.Join(commandDir, action.Path)
		script, err := scripts.Parse(scriptPath)
		if err != nil {
			log.Fatal(err)
		}

		next := scripts.Command{}
		next.Script = script
		next.Arguments = action.Args

		return NewPushCmd(next)
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
