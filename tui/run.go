package tui

import (
	"encoding/json"
	"errors"
	"log"
	"os/exec"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/sunbeam"
	"github.com/skratchdot/open-golang/open"
)

type RunContainer struct {
	width, height int
	command       sunbeam.Command
	input         sunbeam.CommandInput
	embed         Container
}

func NewRunContainer(command sunbeam.Command, input sunbeam.CommandInput) *RunContainer {
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
		items := make([]sunbeam.ListItem, 0)
		for _, row := range rows {
			if row == "" {
				continue
			}
			var item sunbeam.ListItem
			err := json.Unmarshal([]byte(row), &item)
			if err != nil {
				return c, NewErrorCmd("Invalid row: %s", row)
			}
			items = append(items, item)
		}

		c.setEmbed(NewListContainer(c.command.Title, items, c.RunAction))
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

func (c *RunContainer) RunAction(action sunbeam.ScriptAction) tea.Cmd {
	if action.Root == "" {
		action.Root = c.command.Root.String()
	}
	return RunAction(action)
}

func (c *RunContainer) View() string {
	if c.embed == nil {
		return ""
	}
	return c.embed.View()
}

func RunAction(action sunbeam.ScriptAction) tea.Cmd {
	switch action.Type {
	case "push":
		commandMap, ok := sunbeam.ExtensionMap[action.Root]
		if !ok {
			return NewErrorCmd("extension %s does not exists", action.Root)
		}
		command, ok := commandMap[action.Target]
		if !ok {
			return NewErrorCmd("command not found: %s", action.Command)
		}

		input := sunbeam.CommandInput{}
		if action.Params != nil {
			input.Params = action.Params
		} else {
			input.Params = make(map[string]any)
		}

		return NewPushCmd(NewRunContainer(command, input))
	case "exec":
		var cmd *exec.Cmd
		log.Printf("executing command: %s", action.Command)
		if len(action.Command) == 1 {
			cmd = exec.Command(action.Command[0])
		} else {
			cmd = exec.Command(action.Command[0], action.Command[1:]...)
		}
		_, err := cmd.Output()
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return NewErrorCmd("Unable to run cmd: %s", exitError.Stderr)
		}
		return tea.Quit
	case "open":
		err := open.Run(action.Path)
		if err != nil {
			return NewErrorCmd("failed to open file: %s", err)
		}
		return tea.Quit
	case "open-url":
		err := open.Run(action.Url)
		if err != nil {
			return NewErrorCmd("failed to open url: %s", action.Url)
		}
		return tea.Quit
	case "copy":
		err := clipboard.WriteAll(action.Content)
		if err != nil {
			return NewErrorCmd("failed to copy %s to clipboard", err)
		}
		return tea.Quit
	default:
		log.Printf("Unknown action type: %s", action.Type)
		return tea.Quit
	}
}
