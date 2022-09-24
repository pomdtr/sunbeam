package pages

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	commands "github.com/pomdtr/sunbeam/commands"
	"github.com/pomdtr/sunbeam/utils"
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
	title := lipgloss.NewStyle().Render(c.command.Title())
	line := strings.Repeat("─", utils.Max(0, c.width-lipgloss.Width(c.command.Title())))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
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
			c.embed = NewListContainer(c.command.Title(), msg.List, NewActionRunner(c.command))
		case "detail":
			c.embed = NewDetailContainer(msg.Detail, NewActionRunner(c.command))
		}
		c.embed.SetSize(c.width, c.height)
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

func NewActionRunner(command commands.Command) func(commands.ScriptAction) tea.Cmd {
	return func(action commands.ScriptAction) tea.Cmd {
		var cmd tea.Cmd
		callback := func(params any) {
			command.Input.Params = params
			cmd = NewPushCmd(NewCommandContainer(command))
		}
		commands.RunAction(action, callback)
		if cmd != nil {
			return cmd
		}
		return tea.Quit
	}
}
