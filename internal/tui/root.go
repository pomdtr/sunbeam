package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type RootList struct {
	width, height int
	title         string
	history       History
	err           *Detail
	list          *List
	extensions    extensions.ExtensionMap
}

func NewRootList(title string, extensionMap extensions.ExtensionMap) *RootList {
	history, err := LoadHistory(filepath.Join(utils.CacheHome(), "history.json"))
	if err != nil {
		history = History{
			entries: make(map[string]int64),
			path:    filepath.Join(utils.CacheHome(), "history.json"),
		}
	}

	items := make([]types.ListItem, 0)
	for alias, extension := range extensionMap {
		for _, command := range extension.RootCommands() {
			items = append(items, types.ListItem{
				Id:          fmt.Sprintf("extensions/%s/%s", alias, command.Name),
				Title:       command.Title,
				Subtitle:    extension.Title,
				Accessories: []string{alias},
				Actions: []types.Action{
					{
						Title:     "Run",
						Type:      types.ActionTypeRun,
						Extension: alias,
						Command:   command.Name,
					},
				},
			})
		}
	}
	history.Sort(items)
	list := NewList(items...)

	return &RootList{
		title:      title,
		history:    history,
		list:       list,
		extensions: extensionMap,
	}
}

func (c *RootList) Init() tea.Cmd {
	return c.list.Init()
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
	c.list.SetSize(width, height)
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
		case "ctrl+e":
			envFile := filepath.Join(utils.ConfigHome(), "sunbeam.env")
			editCmd := exec.Command("sh", "-c", fmt.Sprintf("$EDITOR %s", envFile))
			return c, tea.ExecProcess(editCmd, func(err error) tea.Msg {
				return err
			})
		}
	case types.Action:
		switch msg.Type {
		case types.ActionTypeRun:
			selection, ok := c.list.Selection()
			if !ok {
				return c, nil
			}

			c.history.entries[selection.Id] = time.Now().Unix()
			if err := c.history.Save(); err != nil {
				return c, c.SetError(err)
			}

			extension, ok := c.extensions[msg.Extension]
			if !ok {
				return c, c.SetError(fmt.Errorf("extension %s not found", msg.Extension))
			}

			command, ok := extension.Command(msg.Command)
			if !ok {
				return c, c.SetError(fmt.Errorf("command %s not found", msg.Command))
			}

			switch command.Mode {
			case types.CommandModePage:
				runner := NewRunner(extension, types.CommandInput{
					Command: command.Name,
					Params:  msg.Params,
				})
				return c, PushPageCmd(runner)
			case types.CommandModeSilent:
				return c, func() tea.Msg {
					if err := extension.Run(types.CommandInput{
						Command: command.Name,
						Params:  msg.Params,
					}); err != nil {
						return err
					}

					return nil
				}
			case types.CommandModeTTY:
				cmd, err := extension.Cmd(types.CommandInput{
					Command: command.Name,
					Params:  msg.Params,
				})

				if err != nil {
					c.err = NewErrorPage(err)
					c.err.SetSize(c.width, c.height)
					return c, c.err.Init()
				}

				return c, tea.ExecProcess(cmd, func(err error) tea.Msg {
					if err != nil {
						return err
					}

					return nil
				})
			}
		case types.ActionTypeCopy:
			return c, func() tea.Msg {
				if err := clipboard.WriteAll(msg.Text); err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			}
		case types.ActionTypeOpen:
			return c, func() tea.Msg {
				if err := utils.OpenWith(msg.Target, msg.App); err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			}
		case types.ActionTypeReload:
			return c, nil
		case types.ActionTypeExit:
			return c, ExitCmd
		default:
			return c, nil
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

type History struct {
	entries map[string]int64
	path    string
}

func (h History) Sort(items []types.ListItem) {
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
