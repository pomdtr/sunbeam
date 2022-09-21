package containers

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	commands "github.com/pomdtr/sunbeam/commands"
)

type ScriptContainer struct {
	width   int
	height  int
	command commands.Command
	spinner spinner.Model
	embed   Container
}

func NewCommandContainer(command commands.Command) *ScriptContainer {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return &ScriptContainer{command: command, spinner: s}
}

func (c *ScriptContainer) SetSize(width, height int) {
	c.width = width
	c.height = height
	if c.embed != nil {
		c.embed.SetSize(width, height-1)
	}
}

func (c *ScriptContainer) Init() tea.Cmd {
	return tea.Batch(c.spinner.Tick, c.fetchItems(c.command))
}

func (c ScriptContainer) fetchItems(command commands.Command) tea.Cmd {
	return func() tea.Msg {
		res, err := command.Run()
		if err != nil {
			return err
		}
		return res
	}
}

func (c *ScriptContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
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
			listView := NewListContainer(msg.List, NewActionRunner(c.command))
			c.embed = listView
		case "detail":
			detailView := NewDetailContainer(msg.Detail, NewActionRunner(c.command))
			c.embed = detailView
		}
		c.embed.SetSize(c.width, c.height-1)
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

func (container *ScriptContainer) View() string {
	title := titleStyle.Render(container.command.Title())

	var content string
	if container.embed == nil {
		label := lipgloss.NewStyle().Padding(1, 2).Render("Loading...")
		content = lipgloss.JoinHorizontal(lipgloss.Center, container.spinner.View(), label)
	} else {
		content = container.embed.View()
	}

	return lipgloss.JoinVertical(lipgloss.Top, title, content)
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
