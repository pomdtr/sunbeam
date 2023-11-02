package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	form          *Form
	extensions    extensions.ExtensionMap
	generator     func() (extensions.ExtensionMap, []types.RootItem, error)
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func cleanString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

func NewRootList(title string, generator func() (extensions.ExtensionMap, []types.RootItem, error)) *RootList {
	history, err := LoadHistory(filepath.Join(utils.CacheHome(), "history.json"))
	if err != nil {
		history = History{
			entries: make(map[string]int64),
			path:    filepath.Join(utils.CacheHome(), "history.json"),
		}
	}

	return &RootList{
		title:     title,
		list:      NewList(),
		history:   history,
		generator: generator,
	}
}

func (c *RootList) Init() tea.Cmd {
	return tea.Batch(c.list.Init(), c.list.SetIsLoading(true), c.Reload)
}

func (c *RootList) Reload() tea.Msg {
	extensionMap, rootItems, err := c.generator()
	if err != nil {
		return err
	}

	items := make([]types.ListItem, 0)
	for _, rootItem := range rootItems {
		extension, ok := extensionMap[rootItem.Extension]
		if !ok {
			continue
		}

		itemID := fmt.Sprintf("%s/%s", cleanString(rootItem.Extension), cleanString(rootItem.Title))

		item := types.ListItem{
			Id:          itemID,
			Title:       rootItem.Title,
			Subtitle:    extension.Title,
			Accessories: []string{rootItem.Extension},
			Actions: []types.Action{
				{
					Title:     "Run",
					Type:      types.ActionTypeRun,
					Extension: rootItem.Extension,
					Command:   rootItem.Command,
					Params:    rootItem.Params,
				},
			},
		}

		if extension.Type == extensions.ExtensionTypeLocal {
			item.Actions = append(item.Actions, types.Action{
				Title:  "Edit Extension",
				Key:    "e",
				Reload: true,
				Type:   types.ActionTypeEdit,
				Target: extension.Entrypoint,
			})
		}

		item.Actions = append(item.Actions, types.Action{
			Title: "Copy Origin",
			Key:   "c",
			Type:  types.ActionTypeCopy,
			Text:  extension.Origin,
			Exit:  true,
		})

		items = append(items, item)
	}

	c.extensions = extensionMap
	return items
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
	case []types.ListItem:
		items := msg
		c.list.SetIsLoading(false)
		c.history.Sort(items)
		c.list.SetItems(items...)
		return c, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+e":
			configFile := filepath.Join(utils.ConfigHome(), "config.json")
			editor := utils.FindEditor()
			editCmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", editor, configFile))
			return c, tea.ExecProcess(editCmd, func(err error) tea.Msg {
				if err != nil {
					return err
				}

				return types.Action{
					Type: types.ActionTypeReload,
				}
			})
		case "ctrl+r":
			c.list.SetIsLoading(true)
			return c, c.Reload
		}
	case types.Action:
		switch msg.Type {
		case types.ActionTypeRun:
			extension, ok := c.extensions[msg.Extension]
			if !ok {
				return c, c.SetError(fmt.Errorf("extension %s not found", msg.Extension))
			}

			command, ok := extension.Command(msg.Command)
			if !ok {
				return c, c.SetError(fmt.Errorf("command %s not found", msg.Command))
			}

			missing := FindMissingParams(command.Params, msg.Params)
			if len(missing) > 0 {
				c.form = NewForm(func(values map[string]any) tea.Msg {
					params := make(map[string]any)
					for k, v := range msg.Params {
						params[k] = v
					}

					for k, v := range values {
						params[k] = v
					}

					return types.Action{
						Title:   msg.Title,
						Type:    types.ActionTypeRun,
						Command: msg.Command,
						Params:  params,
						Exit:    msg.Exit,
						Reload:  msg.Reload,
					}
				}, missing...)

				c.form.SetSize(c.width, c.height)
				return c, tea.Sequence(c.form.Init(), c.form.Focus())
			}
			c.form = nil

			selection, ok := c.list.Selection()
			if !ok {
				return c, nil
			}

			c.history.entries[selection.Id] = time.Now().Unix()
			if err := c.history.Save(); err != nil {
				return c, c.SetError(err)
			}

			switch command.Mode {
			case types.CommandModeList, types.CommandModeDetail:
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
						return PushPageMsg{NewErrorPage(err)}
					}

					return ExitMsg{}
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
						return PushPageMsg{NewErrorPage(err)}
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
		case types.ActionTypeEdit:
			editCmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", utils.FindEditor(), msg.Target))
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

	page, cmd := c.list.Update(msg)
	c.list = page.(*List)

	return c, cmd
}

func (c *RootList) View() string {
	if c.err != nil {
		return c.err.View()
	}
	if c.form != nil {
		return c.form.View()
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
