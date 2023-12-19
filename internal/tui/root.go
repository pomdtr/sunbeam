package tui

import (
	"bytes"
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
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
)

type RootList struct {
	width, height int
	title         string
	err           *Detail
	list          *List
	form          *Form

	config    config.Config
	history   history.History
	generator func() (config.Config, []sunbeam.ListItem, error)
}

type ReloadMsg struct{}

func NewRootList(title string, history history.History, generator func() (config.Config, []sunbeam.ListItem, error)) *RootList {
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
			editCmd := exec.Command("sunbeam", "edit", config.Path)
			return c, tea.ExecProcess(editCmd, func(err error) tea.Msg {
				if err != nil {
					return err
				}

				termenv.DefaultOutput().SetWindowTitle(c.title)
				c.list.Focus()

				return ReloadMsg{}
			})
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
		c.history.Update(selection.Id)
		if err := c.history.Save(); err != nil {
			return c, c.SetError(err)
		}

		switch msg.Type {
		case sunbeam.ActionTypeRun:
			extensionConfig := c.config.Extensions[msg.Run.Extension]
			extension, err := extensions.LoadExtension(extensionConfig.Origin)
			if err != nil {
				return c, c.SetError(fmt.Errorf("failed to load extension: %w", err))
			}

			preferences := extensionConfig.Preferences
			if preferences == nil {
				preferences = make(map[string]any)
			}

			envs, err := ExtractPreferencesFromEnv(msg.Run.Extension, extension)
			if err != nil {
				return c, c.SetError(err)
			}
			for name, value := range envs {
				preferences[name] = value
			}

			missingPreferences := FindMissingPreferences(extension.Manifest.Preferences, preferences)
			for _, preference := range missingPreferences {
				if preference.Optional {
					continue
				}

				c.form = NewForm(func(values map[string]any) tea.Msg {
					if extensionConfig.Preferences == nil {
						extensionConfig.Preferences = make(map[string]any)
					}

					for k, v := range values {
						extensionConfig.Preferences[k] = v
					}

					c.config.Extensions[msg.Run.Extension] = extensionConfig
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

			command, ok := extension.Command(msg.Run.Command)
			if !ok {
				return c, c.SetError(fmt.Errorf("command %s not found", msg.Run.Command))
			}

			missingParams := FindMissingInputs(command.Params, msg.Run.Params)
			for _, param := range missingParams {
				if param.Optional {
					continue
				}

				c.form = NewForm(func(values map[string]any) tea.Msg {
					params := make(map[string]sunbeam.Param)
					for k, v := range msg.Run.Params {
						params[k] = v
					}

					for k, v := range values {
						params[k] = sunbeam.Param{
							Value: v,
						}
					}

					props := msg.Run
					props.Params = params
					return sunbeam.Action{
						Title: msg.Title,
						Type:  sunbeam.ActionTypeRun,
						Run:   props,
					}
				}, missingParams...)

				c.form.SetSize(c.width, c.height)
				return c, c.form.Init()
			}
			c.form = nil

			input := sunbeam.Payload{
				Command:     command.Name,
				Params:      make(map[string]any),
				Preferences: preferences,
			}

			for k, v := range msg.Run.Params {
				input.Params[k] = v.Value
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

					if msg.Run.Exit {
						return ExitMsg{}
					}

					if len(output) > 0 {
						output = bytes.Trim(output, "\n")
						rows := strings.Split(string(output), "\n")
						return ShowNotificationMsg{rows[len(rows)-1]}
					}

					return nil
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

					if msg.Run.Exit {
						return ExitMsg{}
					}

					termenv.DefaultOutput().SetWindowTitle(c.title)
					return c.list.Focus()
				})
			}
		case sunbeam.ActionTypeCopy:
			return c, func() tea.Msg {
				if err := clipboard.WriteAll(msg.Copy.Text); err != nil {
					return err
				}

				if msg.Copy.Exit {
					return ExitMsg{}
				}

				return ShowNotificationMsg{"Copied!"}
			}
		case sunbeam.ActionTypeEdit:
			editCmd := exec.Command("sunbeam", "edit", msg.Edit.Path)
			return c, tea.ExecProcess(editCmd, func(err error) tea.Msg {
				if err != nil {
					return err
				}

				if msg.Edit.Reload {
					return c.Reload()
				}

				if msg.Edit.Exit {
					return ExitMsg{}
				}

				return nil
			})
		case sunbeam.ActionTypeConfig:
			extensionConfig, ok := c.config.Extensions[msg.Config.Extension]
			if !ok {
				return c, c.SetError(fmt.Errorf("extension %s not found", msg.Config.Extension))
			}

			extension, err := extensions.LoadExtension(extensionConfig.Origin)
			if err != nil {
				return c, c.SetError(fmt.Errorf("failed to load extension %s", msg.Config.Extension))
			}

			inputs := make([]sunbeam.Input, 0)
			for _, input := range extension.Manifest.Preferences {
				if preference := extensionConfig.Preferences[input.Name]; preference != nil {
					if err := input.SetDefault(preference); err != nil {
						return c, c.SetError(err)
					}
				}
				input.Optional = false
				inputs = append(inputs, input)
			}

			c.form = NewForm(func(values map[string]any) tea.Msg {
				c.form = nil
				extensionConfig.Preferences = values
				c.config.Extensions[msg.Config.Extension] = extensionConfig
				if err := c.config.Save(); err != nil {
					return err
				}

				return nil
			}, inputs...)
			c.form.SetSize(c.width, c.height)
			return c, c.form.Init()
		case sunbeam.ActionTypeExec:
			cmd := exec.Command("sunbeam", "shell", "-c", msg.Exec.Command)
			cmd.Dir = msg.Exec.Dir
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

			if !msg.Exec.Interactive {
				return c, func() tea.Msg {
					output, err := cmd.Output()
					if err != nil {
						return err
					}

					if msg.Exec.Exit {
						return ExitMsg{}
					}

					if len(output) > 0 {
						output = bytes.Trim(output, "\n")
						rows := strings.Split(string(output), "\n")
						return ShowNotificationMsg{rows[len(rows)-1]}
					}

					return nil
				}
			}

			return c, tea.ExecProcess(cmd, func(err error) tea.Msg {
				if err != nil {
					return err
				}

				if msg.Exec.Exit {
					return ExitMsg{}
				}

				termenv.DefaultOutput().SetWindowTitle(c.title)
				return nil
			})
		case sunbeam.ActionTypeOpen:
			return c, func() tea.Msg {
				if msg.Open.Url != "" {
					if err := utils.Open(msg.Open.Url); err != nil {
						return err
					}

					return ExitMsg{}
				} else if msg.Open.Path != "" {
					if err := utils.Open(fmt.Sprintf("file://%s", msg.Open.Path)); err != nil {
						return err
					}

					return ExitMsg{}
				} else {
					return fmt.Errorf("invalid target")
				}
			}
		case sunbeam.ActionTypeExit:
			return c, ExitCmd
		case sunbeam.ActionTypeReload:
			return c, tea.Sequence(c.err.SetIsLoading(true), c.Reload())
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
