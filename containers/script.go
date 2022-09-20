package containers

import (
	"fmt"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	commands "github.com/pomdtr/sunbeam/commands"
	"github.com/skratchdot/open-golang/open"
)

type ScriptContainer struct {
	width   int
	height  int
	command commands.Command
	Args    []string
	embed   Container
}

func NewScriptContainer(command commands.Command) *ScriptContainer {
	return &ScriptContainer{command: command}
}

func (c *ScriptContainer) SetSize(width, height int) {
	c.width = width
	c.height = height
	if c.embed != nil {
		c.embed.SetSize(width, height-3)
	}
}

func (c *ScriptContainer) Init() tea.Cmd {
	return c.fetchItems(c.command)
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
			listView := NewListContainer(c.command, msg.List.Items)
			listView.SetSize(c.width, c.height-3)
			c.embed = listView
		case "detail":
			detailView := NewDetailContainer(c.command, msg.Detail)
			detailView.SetSize(c.width, c.height-3)
			c.embed = detailView
		}
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
		content = lipgloss.NewStyle().Padding(1, 2).Render("Loading...")
	} else {
		content = container.embed.View()
	}

	return lipgloss.JoinVertical(lipgloss.Top, title, content)
}

func RunAction(command commands.Command, action commands.ScriptAction) (c Container, err error) {
	switch action.Type {
	case "open":
		open.Run(action.Path)
	case "open-url":
		open.Run(action.Url)
	case "copy":
		clipboard.WriteAll(action.Content)
	case "callback":
		c = NewScriptContainer(commands.Command{
			Script: command.Script,
			Args:   command.Args,
			Input: commands.CommandInput{
				Params: action.Params,
			},
		})
	default:
		err = fmt.Errorf("unknown action type: %s", action.Type)
	}
	return
}
