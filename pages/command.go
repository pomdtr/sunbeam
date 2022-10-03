package pages

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/bubbles"
	commands "github.com/pomdtr/sunbeam/commands"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/skratchdot/open-golang/open"
)

type CommandContainer struct {
	width   int
	height  int
	command commands.Command
	spinner spinner.Model
	embed   Page
}

func NewCommandContainer(command commands.Command) *CommandContainer {
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

func (c CommandContainer) fetchItems(command commands.Command) tea.Cmd {
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
	case commands.ScriptResponse:
		switch msg.Type {
		case "list":
			list := msg.List
			if list.Title == "" {
				list.Title = c.command.Title()
			}
			c.embed = NewListContainer(msg.List, c.RunAction)
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
				case "stdin":
					c.command.Form = values
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

var titleStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("62")).
	Foreground(lipgloss.Color("230")).
	Margin(0, 2).
	Padding(0, 1)

func (container *CommandContainer) View() string {
	if container.embed != nil {
		return container.embed.View()
	}

	var loadingIndicator string
	spinner := lipgloss.NewStyle().Padding(0, 2).Render(container.spinner.View())
	label := lipgloss.NewStyle().Render("Loading...")
	loadingIndicator = lipgloss.JoinHorizontal(lipgloss.Center, spinner, label)
	loadingIndicator = lipgloss.NewStyle().Padding(1, 0).Render(loadingIndicator)

	newLines := strings.Repeat("\n", utils.Max(0, container.height-lipgloss.Height(loadingIndicator)-lipgloss.Height(container.footerView())-lipgloss.Height(container.headerView())-1))

	return lipgloss.JoinVertical(lipgloss.Left, container.headerView(), loadingIndicator, newLines, container.footerView())
}

func (c CommandContainer) RunAction(action commands.ScriptAction) tea.Cmd {
	switch action.Type {
	case "callback":
		return c.fetchItems(c.command)
	case "push":
		commandDir := path.Dir(c.command.Url.Path)
		scriptPath := path.Join(commandDir, action.Path)
		script, err := commands.Parse(scriptPath)
		if err != nil {
			log.Fatal(err)
		}

		next := commands.Command{}
		next.Script = script
		next.Arguments = action.Args

		return NewPushCmd(NewCommandContainer(next))
	case "open":
		_ = open.Run(action.Path)
		return tea.Quit
	case "open-url":
		_ = open.Run(action.Url)
		return tea.Quit
	case "copy":
		_ = clipboard.WriteAll(action.Content)
		return tea.Quit
	default:
		log.Printf("Unknown action type: %s", action.Type)
		return tea.Quit
	}

}
