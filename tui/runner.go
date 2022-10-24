package tui

import (
	"encoding/json"
	"strings"

	"github.com/alessio/shellescape"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/api"
)

type RunnerContainer struct {
	width, height int
	command       api.Command
	input         api.CommandInput
	embed         Container
}

func NewRunnerContainer(command api.Command, input api.CommandInput) *RunnerContainer {
	return &RunnerContainer{command: command, input: input, embed: NewLoadingContainer(command.Title)}
}

func (c *RunnerContainer) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.embed.SetSize(width, height)
}

type initMsg struct{}

func (c *RunnerContainer) Init() tea.Cmd {
	missing := c.command.CheckMissingParams(c.input.Params)
	if len(missing) > 0 {
		c.embed = NewFormContainer(c.command.Title, missing)
		c.embed.SetSize(c.width, c.height)
		return c.embed.Init()
	} else {
		return NewSubmitCmd(c.input.Params)
	}
}

type CommandOutput string

func (c *RunnerContainer) SetEmbed(embed Container) tea.Cmd {
	embed.SetSize(c.width, c.height)
	c.embed = embed
	return embed.Init()
}

func (c *RunnerContainer) RunCmd() tea.Cmd {
	return func() tea.Msg {
		output, err := c.command.Run(c.input)
		if err != nil {
			return err
		}
		return CommandOutput(output)
	}
}

func (c *RunnerContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case SubmitMsg:
		for key, value := range msg.values {
			c.input.Params[key] = shellescape.Quote(value)
		}
		cmd := c.SetEmbed(NewLoadingContainer(c.command.Title))
		return c, tea.Batch(cmd, c.RunCmd())
	case ReloadMsg:
		for key, value := range msg.input.Params {
			c.input.Params[key] = shellescape.Quote(value)
		}
		c.input.Query = msg.input.Query
		return c, c.RunCmd()
	case CommandOutput:
		output := string(msg)
		if c.command.Mode != "list" {
			cmd := c.SetEmbed(NewDetailContainer(c.command.Title, output))
			return c, cmd
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

		cmd := c.SetEmbed(NewListContainer(c.command.Title, items, c.input.Query))
		return c, cmd
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

func (c *RunnerContainer) View() string {
	if c.embed == nil {
		return ""
	}
	return c.embed.View()
}
