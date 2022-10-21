package cli

import (
	"encoding/json"
	"strings"

	"github.com/alessio/shellescape"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/commands"
)

type RunContainer struct {
	width, height int
	command       commands.Command
	input         commands.CommandInput
	embed         Container
}

func NewRunContainer(command commands.Command, input commands.CommandInput) *RunContainer {
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
		output = strings.Trim(output, "\n ")
		return CommandOutput(output)
	})
}

func (c *RunContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case SubmitMsg:
		for key, value := range msg.values {
			if value, ok := value.(string); ok {
				c.input.Params[key] = shellescape.Quote(value)
			} else {
				c.input.Params[key] = value
			}
		}
		return c, c.RunCmd()
	case CommandOutput:
		output := string(msg)
		if c.command.Mode != "list" {
			c.setEmbed(NewDetailContainer(c.command.Title, output))
			return c, c.embed.Init()
		}

		rows := strings.Split(output, "\n")
		items := make([]commands.ListItem, len(rows))
		for i, row := range rows {
			var item commands.ListItem
			err := json.Unmarshal([]byte(row), &item)
			if err != nil {
				return c, NewErrorCmd("Invalid row: %s", row)
			}
			items[i] = item
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
