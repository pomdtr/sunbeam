package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/history"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
)

type RootList struct {
	width, height int
	title         string
	err           *Detail
	list          *List

	config    config.Config
	history   *history.History
	generator func() (config.Config, []sunbeam.ListItem, error)
}

type ReloadMsg struct{}

func NewRootList(title string, history *history.History, generator func() (config.Config, []sunbeam.ListItem, error)) *RootList {
	return &RootList{
		title:     title,
		history:   history,
		generator: generator,
	}
}

func (c *RootList) Init() tea.Cmd {
	termenv.DefaultOutput().SetWindowTitle(c.title)
	return c.Reload()
}

func (c *RootList) Reload() tea.Cmd {
	cfg, rootItems, err := c.generator()
	if err != nil {
		return c.SetError(err)
	}

	c.config = cfg
	if c.history != nil {
		c.history.Sort(rootItems)
	}
	if c.list != nil {
		c.list.SetIsLoading(false)
		c.list.SetItems(rootItems...)
		return nil
	} else {
		c.list = NewList(rootItems...)
		c.list.SetEmptyText("No items")
		c.list.SetSize(c.width, c.height)

		return c.list.Init()
	}
}

func (c *RootList) Focus() tea.Cmd {
	termenv.DefaultOutput().SetWindowTitle(c.title)
	return c.list.Focus()
}

func (c *RootList) Blur() tea.Cmd {
	return c.list.SetIsLoading(false)
}

func (c *RootList) SetSize(width, height int) {
	c.width, c.height = width, height
	if c.err != nil {
		c.err.SetSize(width, height)
	}

	if c.list != nil {
		c.list.SetSize(width, height)
	}
}

func (c *RootList) SetError(err error) tea.Cmd {
	c.err = NewErrorPage(err)
	c.err.SetSize(c.width, c.height)
	return func() tea.Msg {
		return err
	}
}

func (c *RootList) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+r":
			return c, tea.Batch(c.list.SetIsLoading(true), c.Reload())
		}
	case ReloadMsg:
		return c, tea.Batch(c.list.SetIsLoading(true), c.Reload())
	case sunbeam.Action:
		selection, ok := c.list.Selection()
		if !ok {
			return c, nil
		}
		if c.history != nil {
			c.history.Update(selection.Id)
			if err := c.history.Save(); err != nil {
				return c, c.SetError(err)
			}
		}

		extensionConfig := c.config.Extensions[msg.Extension]
		extension, err := extensions.LoadExtension(extensionConfig.Origin)
		if err != nil {
			return c, c.SetError(fmt.Errorf("failed to load extension: %w", err))
		}

		extension.Env = c.config.Env
		for k, v := range extensionConfig.Env {
			extension.Env[k] = v
		}

		command, ok := extension.Command(msg.Command)
		if !ok {
			return c, c.SetError(fmt.Errorf("command %s not found", msg.Command))
		}

		input := sunbeam.Payload{
			Command: command.Name,
			Params:  msg.Params,
		}

		switch command.Mode {
		case sunbeam.CommandModeSearch, sunbeam.CommandModeFilter, sunbeam.CommandModeDetail:
			runner := NewRunner(extension, input)
			return c, PushPageCmd(runner)
		case sunbeam.CommandModeSilent:
			return c, func() tea.Msg {
				output, err := extension.Output(input)
				if err != nil {
					return PushPageMsg{NewErrorPage(err)}
				}

				if len(output) > 0 {
					var action sunbeam.Action
					if err := json.Unmarshal(output, &action); err != nil {
						return PushPageMsg{NewErrorPage(err)}
					}

					return action
				}

				if msg.Reload {
					return ReloadMsg{}
				}

				return ExitMsg{}
			}
		case sunbeam.CommandModeTTY:
			cmd, err := extension.Cmd(input)

			if err != nil {
				c.err = NewErrorPage(err)
				c.err.SetSize(c.width, c.height)
				return c, c.err.Init()
			}

			return c, tea.ExecProcess(cmd, func(err error) tea.Msg {
				if err != nil {
					return PushPageMsg{NewErrorPage(err)}
				}

				if msg.Reload {
					return ReloadMsg{}
				}

				return ExitMsg{}
			})
		}
	case error:
		c.err = NewErrorPage(msg)
		c.err.SetSize(c.width, c.height)
		return c, c.err.Init()
	}

	if c.err != nil {
		page, cmd := c.err.Update(msg)
		c.err = page.(*Detail)
		return c, cmd
	}

	if c.list != nil {
		page, cmd := c.list.Update(msg)
		c.list = page.(*List)
		return c, cmd
	}

	return c, nil
}

func (c *RootList) View() string {
	if c.err != nil {
		return c.err.View()
	}

	if c.list != nil {
		return c.list.View()
	}

	return ""
}

type History struct {
	entries map[string]int64
	path    string
}

func (h History) Sort(items []sunbeam.ListItem) {
	sort.SliceStable(items, func(i, j int) bool {
		keyI := items[i].Id
		keyJ := items[j].Id

		return h.entries[keyI] > h.entries[keyJ]
	})
}

func LoadHistory(fp string) (History, error) {
	f, err := os.Open(fp)
	if err != nil {
		return History{}, err
	}

	var entries map[string]int64
	if err := json.NewDecoder(f).Decode(&entries); err != nil {
		return History{}, err
	}

	return History{
		entries: entries,
		path:    fp,
	}, nil
}

func (h History) Save() error {
	f, err := os.OpenFile(h.path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(h.path), 0755); err != nil {
			return err
		}

		f, err = os.Create(h.path)
		if err != nil {
			return err
		}
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(h.entries); err != nil {
		return err
	}

	return nil
}
