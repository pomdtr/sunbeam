package tui

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/alessio/shellescape"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sunbeamlauncher/sunbeam/app"
	"github.com/sunbeamlauncher/sunbeam/utils"
)

type Config struct {
	Height int

	RootItems []app.RootItem `yaml:"rootItems"`
}

type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type Model struct {
	width, height int
	exitCmd       *exec.Cmd

	config       *Config
	root         Page
	pages        []Page
	extensionMap map[string]app.Extension

	hidden bool
}

func NewModel(config *Config, extensions ...app.Extension) *Model {
	extensionMap := make(map[string]app.Extension)
	rootItems := make([]app.RootItem, 0)
	for _, extension := range extensions {
		extensionMap[extension.Name] = extension
		rootItems = append(rootItems, extension.RootItems...)
	}
	for _, rootItem := range config.RootItems {
		if _, ok := extensionMap[rootItem.Extension]; !ok {
			continue
		}
		rootItems = append(rootItems, rootItem)
	}
	rootList := NewRootList(rootItems...)

	return &Model{extensionMap: extensionMap, root: rootList, config: config}
}

func (m *Model) SetRoot(root Page) {
	m.root = root
}

func (m *Model) Init() tea.Cmd {
	return m.root.Init()
}

func (m Model) IsFullScreen() bool {
	return m.config.Height == 0
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.hidden = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case OpenMsg:
		m.hidden = true
		
		_, err := exec.Command("systemd-run", "--scope", "--user", "xdg-open", msg.Target).CombinedOutput()
		if err != nil {
			return m, NewErrorCmd(err)
		}
		m.hidden = true
		return m, tea.Quit


	case CopyTextMsg:
		m.hidden = true
		return m, func() tea.Msg {
			err := clipboard.WriteAll(msg.Text)
			if err != nil {
				return err
			}

			return tea.Quit()
		}
	case ShowPrefMsg:
		extension, ok := m.extensionMap[msg.Extension]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("extension %s not found", msg.Extension))
		}
		script, ok := extension.Scripts[msg.Script]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("script %s not found", msg.Script))
		}

		pref := NewPreferenceForm(extension, script)

		cmd := m.Push(pref)

		return m, cmd
	case RunScriptMsg:
		extension, ok := m.extensionMap[msg.Extension]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("extension %s not found", msg.Extension))
		}

		if len(extension.Requirements) > 0 {
			for _, requirement := range extension.Requirements {
				if !requirement.Check() {
					container := NewDetail("Requirement not met")
					container.content = fmt.Sprintf("requirement %s not met.\nHomepage: %s", requirement.Which, requirement.HomePage)
					return m, NewPushCmd(container)
				}
			}
		}

		script, ok := extension.Scripts[msg.Script]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("script %s not found", msg.Script))
		}

		runner := NewScriptRunner(extension, script, msg.With)
		runner.OnSuccessCmd = msg.OnSuccessCmd()
		cmd := m.Push(runner)
		return m, cmd
	case ExecCommandMsg:
		command := exec.Command("sh", "-c", msg.Command)
		command.Dir = msg.Directory
		command.Env = os.Environ()
		command.Env = append(command.Env, msg.Env...)

		if msg.OnSuccess == "" {
			m.exitCmd = command
			m.hidden = true
			return m, tea.Quit
		}

		m.hidden = true
		return m, tea.ExecProcess(command, func(err error) tea.Msg {
			if err != nil {
				return showMsg{
					cmd: NewErrorCmd(err),
				}
			}

			return showMsg{
				cmd: msg.OnSuccessCmd(),
			}
		})
	case showMsg:
		m.hidden = false
		return m, msg.cmd
	case pushMsg:
		m.hidden = false
		cmd := m.Push(msg.container)
		return m, cmd
	case popMsg:
		if len(m.pages) == 0 {
			return m, tea.Quit
		} else {
			m.Pop()
			return m, nil
		}
	case error:
		m.hidden = false
		detail := NewDetail("Error")
		detail.SetSize(m.width, m.pageHeight())
		detail.SetContent(msg.Error())

		if len(m.pages) == 0 {
			m.root = detail
		} else {
			m.pages[len(m.pages)-1] = detail
		}

		return m, detail.Init()
	}

	// Update the current page
	var cmd tea.Cmd

	if len(m.pages) == 0 {
		m.root, cmd = m.root.Update(msg)
	} else {
		currentPageIdx := len(m.pages) - 1
		m.pages[currentPageIdx], cmd = m.pages[currentPageIdx].Update(msg)
	}

	return m, cmd
}

type ShowPrefMsg struct {
	Extension string
	Script    string
}

func (m *Model) View() string {
	if m.hidden {
		return ""
	}

	if len(m.pages) > 0 {
		currentPage := m.pages[len(m.pages)-1]
		return currentPage.View()
	}

	return m.root.View()

}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	m.root.SetSize(width, m.pageHeight())
	for _, page := range m.pages {
		page.SetSize(m.width, m.pageHeight())
	}
}

func (m *Model) pageHeight() int {
	if m.config.Height > 0 {
		return utils.Min(m.config.Height, m.height)
	} else {
		return m.height
	}
}

type popMsg struct{}

type showMsg struct {
	cmd tea.Cmd
}

func PopCmd() tea.Msg {
	return popMsg{}
}

type pushMsg struct {
	container Page
}

func NewPushCmd(c Page) tea.Cmd {
	return func() tea.Msg {
		return pushMsg{c}
	}
}

func (m *Model) Push(page Page) tea.Cmd {
	page.SetSize(m.width, m.pageHeight())
	m.pages = append(m.pages, page)
	return page.Init()
}

func (m *Model) Pop() {
	if len(m.pages) > 0 {
		m.pages = m.pages[:len(m.pages)-1]
	}
}

func toShellCommand(rootItem app.RootItem) string {
	args := []string{"sunbeam", "run", rootItem.Extension, rootItem.Script}
	for param, value := range rootItem.With {
		switch value := value.(type) {
		case string:
			value = shellescape.Quote(value)
			args = append(args, fmt.Sprintf("--%s=%s", param, value))
		case bool:
			if !value {
				continue
			}
			args = append(args, fmt.Sprintf("--%s", param))
		}
	}
	return strings.Join(args, " ")
}

func loadHistory(historyPath string) map[string]int64 {
	history := make(map[string]int64)
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return history
	}

	json.Unmarshal(data, &history)
	return history
}

func NewRootList(rootItems ...app.RootItem) Page {
	stateDir := path.Join(os.Getenv("HOME"), ".local", "state", "sunbeam")
	historyPath := path.Join(stateDir, "history.json")

	history := loadHistory(historyPath)

	list := NewList("Sunbeam")

	list.filter.Less = func(i, j FilterItem) bool {
		iValue, ok := history[i.ID()]
		if !ok {
			iValue = 0
		}
		jValue, ok := history[j.ID()]
		if !ok {
			jValue = 0
		}

		return iValue > jValue
	}

	listItems := make([]ListItem, 0)
	for _, rootItem := range rootItems {
		rootItem := rootItem
		with := make(map[string]app.ScriptInputWithValue)
		itemShellCommand := toShellCommand(rootItem)
		for key, value := range rootItem.With {
			with[key] = app.ScriptInputWithValue{Value: value}
		}
		listItems = append(listItems, ListItem{
			Id:       itemShellCommand,
			Title:    rootItem.Title,
			Subtitle: rootItem.Subtitle,
			Actions: []Action{
				{
					Title:    "Run Script",
					Shortcut: "enter",
					Cmd: func() tea.Msg {
						history[itemShellCommand] = time.Now().Unix()
						if _, err := os.Stat(stateDir); os.IsNotExist(err) {
							os.MkdirAll(stateDir, 0755)
						}

						data, _ := json.Marshal(history)
						os.WriteFile(historyPath, data, 0644)

						return RunScriptMsg{
							Extension: rootItem.Extension,
							Script:    rootItem.Script,
							With:      with,
						}
					},
				}, {
					Title:    "Show Preferences",
					Shortcut: "ctrl+p",
					Cmd: func() tea.Msg {
						return ShowPrefMsg{
							Extension: rootItem.Extension,
							Script:    rootItem.Script,
						}
					},
				},
				{
					Title:    "Copy as Shell Command",
					Shortcut: "ctrl+y",
					Cmd:      NewCopyTextCmd(itemShellCommand),
				},
			},
		})
	}

	list.SetItems(listItems)

	return list
}

func Draw(model *Model) (err error) {
	// Log to a file
	if env := os.Getenv("SUNBEAM_LOG_FILE"); env != "" {
		f, err := tea.LogToFile(env, "debug")
		if err != nil {
			log.Fatalf("could not open log file: %v", err)
		}
		defer f.Close()
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		logDir := path.Join(home, ".local", "state", "sunbeam")
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			err = os.MkdirAll(path.Join(home, ".local", "state", "sunbeam"), 0755)
			if err != nil {
				return err
			}
		}
		tea.LogToFile(path.Join(logDir, "sunbeam.log"), "")
	}

	var p *tea.Program
	if model.IsFullScreen() {
		p = tea.NewProgram(model, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(model)
	}

	m, err := p.Run()
	if err != nil {
		return err
	}

	model = m.(*Model)

	if exitCmd := model.exitCmd; exitCmd != nil {
		exitCmd.Stderr = os.Stderr
		exitCmd.Stdout = os.Stdout
		exitCmd.Stdin = os.Stdin

		exitCmd.Run()
	}

	return nil
}
