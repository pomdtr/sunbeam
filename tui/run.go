package tui

import (
	"encoding/json"
	"strings"

	"github.com/alessio/shellescape"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/api"
)

type RunContainer struct {
	width, height int
	command       api.Command
	input         api.CommandInput
	embed         Container
}

func NewRunContainer(command api.Command, input api.CommandInput) *RunContainer {
	return &RunContainer{command: command, input: input, embed: NewLoadingContainer(command.Title)}
}

func (c *RunContainer) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.embed.SetSize(width, height)
}

type initMsg struct{}

func (c *RunContainer) Init() tea.Cmd {
	missing := c.command.CheckMissingParams(c.input.Params)
	if len(missing) > 0 {
		c.embed = NewFormContainer(c.command.Title, missing)
		c.embed.SetSize(c.width, c.height)
		return c.embed.Init()
	} else {
		return c.RunCmd()
	}
}

type CommandOutput string

func (c *RunContainer) setEmbed(embed Container) {
	embed.SetSize(c.width, c.height)
	c.embed = embed
}

func (c *RunContainer) RunCmd() tea.Cmd {
	c.embed = NewLoadingContainer(c.command.Title)
	c.embed.SetSize(c.width, c.height)
	return tea.Batch(c.embed.Init(), func() tea.Msg {
		output, err := c.command.Run(c.input)
		if err != nil {
			return err
		}
		return CommandOutput(output)
	})
}

func (c *RunContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case SubmitMsg:
		for key, value := range msg.values {
			c.input.Params[key] = shellescape.Quote(value)
		}
		return c, c.RunCmd()
	case CommandOutput:
		output := string(msg)
		if c.command.Mode != "list" {
			c.setEmbed(NewDetailContainer(c.command.Title, output))
			return c, c.embed.Init()
		}

		rows := strings.Split(output, "\n")
		items := make([]api.ListItem, 0)
		for _, row := range rows {
			if row == "" {
				continue
			}
			var item api.ListItem
			err := json.Unmarshal([]byte(row), &item)
			if err != nil {
				return c, NewErrorCmd("Invalid row: %s", row)
			}
			items = append(items, item)
		}

		c.setEmbed(NewListContainer(c.command.Title, items))
		return c, c.embed.Init()
	case error:
		e := NewDetailContainer("Error", msg.Error())
		e.SetSize(c.width, c.height)
		c.embed = e
		return c, e.Init()
	}

	var cmd tea.Cmd
	c.embed, cmd = c.embed.Update(msg)
	return c, cmd

}

func (c *RunContainer) View() string {
	if c.embed == nil {
		return ""
	}
	return c.embed.View()
}
