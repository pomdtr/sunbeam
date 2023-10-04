package tui

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/shlex"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type RootList struct {
	w, h       int
	err        *Detail
	list       *List
	extensions map[string]Extension
	OnSelect   func(id string)
}

func NewRootList(extensions map[string]Extension, rootItems map[string]string, history map[string]int64) *RootList {
	items := make([]types.ListItem, 0)
	for alias, extension := range extensions {
		for _, command := range extension.Commands {
			if !isRootCommand(command) {
				continue
			}

			shellCommand := fmt.Sprintf("sunbeam %s %s", alias, command.Name)
			items = append(items, types.ListItem{
				Id:          shellCommand,
				Title:       command.Title,
				Subtitle:    extension.Title,
				Accessories: []string{shellCommand},
				Actions: []types.Action{
					{
						Title: "Run Command",
						OnAction: types.Command{
							Type:      types.CommandTypeRun,
							Extension: alias,
							Command:   command.Name,
						},
					},
					{
						Title: "Copy Command",
						OnAction: types.Command{
							Type: types.CommandTypeCopy,
							Text: shellCommand,
							Exit: true,
						},
					},
				},
			})
		}
	}

	for title, shellCommand := range rootItems {
		args, err := shlex.Split(shellCommand)
		if err != nil {
			return nil
		}

		if len(args) == 0 {
			continue
		}

		if args[0] == "sunbeam" {
			ref, err := ExtractCommand(args[1:])
			if err != nil {
				continue
			}

			extension, ok := extensions[ref.Alias]
			if !ok {
				continue
			}

			items = append(items, types.ListItem{
				Id:          shellCommand,
				Title:       title,
				Subtitle:    extension.Title,
				Accessories: []string{shellCommand},
				Actions: []types.Action{
					{
						Title: "Run Command",
						OnAction: types.Command{
							Type:      types.CommandTypeRun,
							Extension: ref.Alias,
							Command:   ref.Command,
							Params:    ref.Params,
						},
					},
					{
						Title: "Copy Command",
						OnAction: types.Command{
							Type: types.CommandTypeCopy,
							Text: shellCommand,
							Exit: true,
						},
					},
				},
			})
		} else {
			items = append(items, types.ListItem{
				Id:          shellCommand,
				Title:       title,
				Subtitle:    "Oneliner",
				Accessories: []string{shellCommand},
				Actions: []types.Action{
					{
						Title: "Run Command",
						OnAction: types.Command{
							Type: types.CommandTypeExec,
							Args: args,
						},
					},
				},
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		timestampA, ok := history[items[i].Id]
		if !ok {
			return false
		}

		timestampB, ok := history[items[j].Id]
		if !ok {
			return true
		}

		return timestampA > timestampB
	})

	return &RootList{
		list:       NewList(items...),
		extensions: extensions,
	}
}

func (c *RootList) Init() tea.Cmd {
	return tea.Batch(c.list.Init())
}

func (c *RootList) Focus() tea.Cmd {
	termOutput.SetWindowTitle("Sunbeam")
	return c.list.Focus()
}

func (c *RootList) Blur() tea.Cmd {
	return c.list.SetIsLoading(false)
}

func (c *RootList) SetSize(width, height int) {
	c.w, c.h = width, height
	if c.err != nil {
		c.err.SetSize(width, height)
	}
	c.list.SetSize(width, height)
}

func (c *RootList) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			item, ok := c.list.Selection()
			if !ok {
				break
			}

			if c.OnSelect != nil {
				c.OnSelect(item.Id)
			}
		}
	case error:
		c.err = NewErrorPage(msg)
		c.err.SetSize(c.w, c.h)
		return c, c.err.Init()
	case types.Command:
		if msg.Type != types.CommandTypeRun {
			switch msg.Type {
			case types.CommandTypeCopy:
				return c, func() tea.Msg {
					if err := clipboard.WriteAll(msg.Text); err != nil {
						return err
					}

					if msg.Exit {
						return ExitMsg{}
					}

					return nil
				}

			case types.CommandTypeOpen:
				return c, func() tea.Msg {
					if err := utils.OpenWith(msg.Target, msg.App); err != nil {
						return err
					}

					if msg.Exit {
						return ExitMsg{}
					}

					return nil
				}
			case types.CommandTypeExec:
				return c, tea.ExecProcess(exec.Command(msg.Args[0], msg.Args[1:]...), func(err error) tea.Msg {
					if err != nil {
						return err
					}

					return ExitMsg{}
				})
			case types.CommandTypeExit:
				return c, ExitCmd
			default:
				return c, nil
			}
		}

		return c, func() tea.Msg {
			extension, ok := c.extensions[msg.Extension]
			if !ok {
				return fmt.Errorf("extension %s not found", msg.Extension)
			}

			command, ok := extension.Command(msg.Command)
			if !ok {
				return fmt.Errorf("command %s not found", msg.Command)
			}

			if command.Mode == types.CommandModeView {
				return PushPageMsg{NewRunner(extension, command, msg.Params)}
			}

			out, err := extension.Run(CommandInput{
				Command: msg.Command,
				Params:  msg.Params,
			})
			if err != nil {
				return err
			}

			if len(out) == 0 {
				return ExitMsg{}
			}

			outputCommand := types.Command{}
			if err := json.Unmarshal(out, &outputCommand); err != nil {
				return err
			}

			return outputCommand
		}
	}

	if c.err != nil {
		page, cmd := c.err.Update(msg)
		c.err = page.(*Detail)
		return c, cmd
	}

	page, cmd := c.list.Update(msg)
	c.list = page.(*List)

	return c, cmd
}

func (c *RootList) View() string {
	if c.err != nil {
		return c.err.View()
	}
	return c.list.View()
}

func isRootCommand(command types.CommandSpec) bool {
	for _, param := range command.Params {
		if !param.Optional {
			return false
		}
	}

	return true
}

type CommandRef struct {
	Alias   string
	Command string
	Params  map[string]any
}

func ExtractCommand(args []string) (CommandRef, error) {
	var ref CommandRef
	if len(args) == 0 {
		return ref, fmt.Errorf("no extension specified")
	}

	ref.Alias = args[0]
	args = args[1:]

	if len(args) == 0 {
		return ref, fmt.Errorf("no command specified")
	}

	ref.Command = args[0]
	args = args[1:]

	if len(args) == 0 {
		return ref, nil
	}

	ref.Params = make(map[string]any)

	for len(args) > 0 {
		if !strings.HasPrefix(args[0], "--") {
			return ref, fmt.Errorf("invalid argument: %s", args[0])
		}

		arg := strings.TrimPrefix(args[0], "--")

		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			ref.Params[parts[0]] = parts[1]
			args = args[1:]
			continue
		}

		if len(args) == 1 {
			ref.Params[arg] = true
			args = args[1:]
			continue
		}

		if strings.HasPrefix(args[1], "--") {
			ref.Params[arg] = true
			args = args[1:]
			continue
		}

		ref.Params[arg] = args[1]
		args = args[2:]
	}

	return ref, nil
}
