package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/history"
	"github.com/pomdtr/sunbeam/internal/types"
	"github.com/pomdtr/sunbeam/internal/utils"
)

type RootList struct {
	width, height int
	title         string
	err           *Detail
	list          *List
	form          *Form

	config    config.Config
	history   history.History
	generator func() (config.Config, []types.ListItem, error)
}

func NewRootList(title string, history history.History, generator func() (config.Config, []types.ListItem, error)) *RootList {
	return &RootList{
		title:     title,
		history:   history,
		generator: generator,
	}
}

func (c *RootList) Init() tea.Cmd {
	termenv.DefaultOutput().SetWindowTitle(c.title)
	return c.Reload
}

func (c *RootList) Reload() tea.Msg {
	cfg, rootItems, err := c.generator()
	if err != nil {
		return err
	}

	c.config = cfg
	c.history.Sort(rootItems)
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
	if c.form != nil {
		c.form.SetSize(width, height)
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
		case "esc":
			if c.form != nil {
				c.form = nil
				return c, c.list.Focus()
			}
		case "ctrl+s":
			if c.form != nil {
				break
			}

			shellCmd := exec.Command(utils.FindShell(), "-li")
			return c, tea.ExecProcess(shellCmd, func(err error) tea.Msg {
				termenv.DefaultOutput().SetWindowTitle(c.title)
				if err != nil {
					return err
				}

				return c.list.Focus()
			})
		case "ctrl+e":
			if c.form != nil {
				break
			}
			editCmd := exec.Command("sunbeam", "edit", config.Path)
			return c, tea.ExecProcess(editCmd, func(err error) tea.Msg {
				if err != nil {
					return err
				}

				termenv.DefaultOutput().SetWindowTitle(c.title)
				c.list.Focus()
				return types.Action{
					Type: types.ActionTypeReload,
				}
			})
		case "ctrl+r":
			return c, tea.Batch(c.list.SetIsLoading(true), c.Reload)
		}
	case types.Action:
		selection, ok := c.list.Selection()
		if !ok {
			return c, nil
		}
		c.history.Update(selection.Id)
		if err := c.history.Save(); err != nil {
			return c, c.SetError(err)
		}

		switch msg.Type {
		case types.ActionTypeRun:
			extensionConfig := c.config.Extensions[msg.Extension]
			extension, err := extensions.LoadExtension(extensionConfig.Origin)
			if err != nil {
				return c, c.SetError(fmt.Errorf("failed to load extension: %w", err))
			}

			preferences := extensionConfig.Preferences
			if preferences == nil {
				preferences = make(map[string]any)
			}

			envs, err := ExtractPreferencesFromEnv(msg.Extension, extension)
			if err != nil {
				return c, c.SetError(err)
			}
			for name, value := range envs {
				preferences[name] = value
			}

			missingPreferences := FindMissingPreferences(extension.Manifest.Preferences, preferences)
			for _, preference := range missingPreferences {
				if !preference.Required {
					continue
				}

				c.form = NewForm(func(values map[string]any) tea.Msg {
					if extensionConfig.Preferences == nil {
						extensionConfig.Preferences = make(map[string]any)
					}

					for k, v := range values {
						extensionConfig.Preferences[k] = v
					}

					c.config.Extensions[msg.Extension] = extensionConfig
					if config.Path == "" {
						return msg
					}

					if err := c.config.Save(); err != nil {
						return err
					}

					return msg
				}, missingPreferences...)

				c.form.SetSize(c.width, c.height)
				return c, c.form.Init()
			}

			command, ok := extension.Command(msg.Command)
			if !ok {
				return c, c.SetError(fmt.Errorf("command %s not found", msg.Command))
			}

			missingParams := FindMissingInputs(command.Params, msg.Params)
			for _, param := range missingParams {
				if !param.Required {
					continue
				}

				c.form = NewForm(func(values map[string]any) tea.Msg {
					params := make(map[string]types.Param)
					for k, v := range msg.Params {
						params[k] = v
					}

					for k, v := range values {
						params[k] = types.Param{
							Value: v,
						}
					}

					return types.Action{
						Title:     msg.Title,
						Type:      types.ActionTypeRun,
						Extension: msg.Extension,
						Command:   msg.Command,
						Params:    params,
						Exit:      msg.Exit,
						Reload:    msg.Reload,
					}
				}, missingParams...)

				c.form.SetSize(c.width, c.height)
				return c, c.form.Init()
			}
			c.form = nil

			input := types.Payload{
				Command:     command.Name,
				Params:      make(map[string]any),
				Preferences: preferences,
			}

			for k, v := range msg.Params {
				input.Params[k] = v.Value
			}

			switch command.Mode {
			case types.CommandModeSearch, types.CommandModeFilter, types.CommandModeDetail:
				runner := NewRunner(extension, input)
				return c, PushPageCmd(runner)
			case types.CommandModeSilent:
				return c, func() tea.Msg {
					if err := extension.Run(input); err != nil {
						return PushPageMsg{NewErrorPage(err)}
					}

					if msg.Exit {
						return ExitMsg{}
					}

					return nil
				}
			case types.CommandModeTTY:
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

					if msg.Exit {
						return ExitMsg{}
					}

					termenv.DefaultOutput().SetWindowTitle(c.title)
					return c.list.Focus()
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

				return ShowNotificationMsg{"Copied!"}
			}
		case types.ActionTypeEdit:
			editCmd := exec.Command("sunbeam", "edit", msg.Path)
			return c, tea.ExecProcess(editCmd, func(err error) tea.Msg {
				if err != nil {
					return err
				}

				if msg.Reload {
					return c.Reload()
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			})
		case types.ActionTypeConfig:
			extensionConfig, ok := c.config.Extensions[msg.Extension]
			if !ok {
				return c, c.SetError(fmt.Errorf("extension %s not found", msg.Extension))
			}

			extension, err := extensions.LoadExtension(extensionConfig.Origin)
			if err != nil {
				return c, c.SetError(fmt.Errorf("failed to load extension %s", msg.Extension))
			}

			inputs := make([]types.Input, 0)
			for _, input := range extension.Manifest.Preferences {
				input.Default = extensionConfig.Preferences[input.Name]
				input.Required = true
				inputs = append(inputs, input)
			}

			c.form = NewForm(func(values map[string]any) tea.Msg {
				c.form = nil
				extensionConfig.Preferences = values
				c.config.Extensions[msg.Extension] = extensionConfig
				if err := c.config.Save(); err != nil {
					return err
				}

				return nil
			}, inputs...)
			c.form.SetSize(c.width, c.height)
			return c, c.form.Init()
		case types.ActionTypeExec:
			cmd := exec.Command("sh", "-c", msg.Command)
			cmd.Dir = msg.Dir
			if strings.HasPrefix(cmd.Dir, "~") {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return c, c.SetError(err)
				}

				cmd.Dir = filepath.Join(homeDir, strings.TrimPrefix(cmd.Dir, "~"))
			}

			if !filepath.IsAbs(cmd.Dir) {
				wd, err := os.Getwd()
				if err != nil {
					return c, c.SetError(err)
				}

				cmd.Dir = filepath.Join(wd, cmd.Dir)
			}

			return c, tea.ExecProcess(cmd, func(err error) tea.Msg {
				if err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				termenv.DefaultOutput().SetWindowTitle(c.title)
				return nil
			})
		case types.ActionTypeOpen:
			return c, func() tea.Msg {
				if msg.Url != "" {
					if err := utils.Open(msg.Url); err != nil {
						return err
					}

					return ExitMsg{}
				} else if msg.Path != "" {
					if err := utils.Open(fmt.Sprintf("file://%s", msg.Path)); err != nil {
						return err
					}

					return ExitMsg{}
				} else {
					return fmt.Errorf("invalid target")
				}
			}
		case types.ActionTypeReload:
			return c, tea.Sequence(c.list.SetIsLoading(true), c.Reload)
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

	if c.form != nil {
		page, cmd := c.form.Update(msg)
		c.form = page.(*Form)
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
	if c.form != nil {
		return c.form.View()
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
