package tui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/browser"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/skratchdot/open-golang/open"
)

type Config struct {
	Height         int
	AccentColor    string
	CopyCommand    string
	OpenUrlCommand string

	RootItems []app.RootItem `yaml:"rootItems"`
}

type Container interface {
	Init() tea.Cmd
	Update(tea.Msg) (Container, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type RootModel struct {
	width, height int
	exit          bool
	config        Config

	pages []Container

	hidden bool
}

func NewRootModel(rootPage Container, config Config) *RootModel {
	return &RootModel{pages: []Container{rootPage}, config: config}
}

func (m *RootModel) Init() tea.Cmd {
	return m.pages[0].Init()
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.exit = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case CopyTextMsg:
		var err error
		if m.config.CopyCommand != "" {
			cmd := exec.Command("sh", "-c", m.config.CopyCommand)
			cmd.Stdin = strings.NewReader(msg.Text)
			err = cmd.Run()
		} else {
			err = clipboard.WriteAll(msg.Text)
		}
		if err != nil {
			return m, NewErrorCmd(err)
		}
		return m, tea.Quit
	case OpenUrlMsg:
		var err error
		if m.config.OpenUrlCommand != "" {
			cmd := exec.Command("sh", "-c", m.config.OpenUrlCommand)
			cmd.Stdin = strings.NewReader(msg.Url)
			err = cmd.Run()
		} else {
			err = browser.OpenURL(msg.Url)
		}
		if err != nil {
			return m, NewErrorCmd(err)
		}
		return m, tea.Quit
	case OpenPathMsg:
		var err error

		if msg.Application != "" {
			err = open.RunWith(msg.Path, msg.Application)
		} else {
			err = open.Run(msg.Path)
		}
		if err != nil {
			return m, NewErrorCmd(err)
		}
		return m, tea.Quit
	case RunScriptMsg:
		extension, ok := app.Sunbeam.Extensions[msg.Extension]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("extension %s not found", msg.Extension))
		}

		if len(extension.Requirements) > 0 {
			for _, requirement := range extension.Requirements {
				if !requirement.Check() {
					container := NewDetail("Requirement not met")
					container.SetContent(fmt.Sprintf("requirement %s not met.\nHomepage: %s", requirement.Which, requirement.HomePage))
					return m, NewPushCmd(container)
				}
			}
		}

		script, ok := extension.Scripts[msg.Script]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("script %s not found", msg.Script))
		}

		runner := NewScriptRunner(extension, script, msg.With)
		cmd := m.Push(runner)
		return m, cmd
	case ExecCommandMsg:
		command := exec.Command("sh", "-c", fmt.Sprintf("%s; clear", msg.Command))
		if msg.Silent {
			err := command.Run()
			if err != nil {
				detail := NewDetail(command.String())
				detail.SetContent(err.Error())
				return m, NewPushCmd(detail)
			}

			return m, msg.OnSuccessCmd
		}

		m.hidden = true
		return m, tea.ExecProcess(command, func(err error) tea.Msg {
			if err != nil {
				log.Println("ERROR")
				detail := NewDetail(command.String())
				detail.SetContent(err.Error())
				return showMsg{
					cmd: NewErrorCmd(err),
				}
			}

			return showMsg{
				cmd: msg.OnSuccessCmd,
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
		if len(m.pages) == 1 {
			return m, tea.Quit
		} else {
			m.Pop()
			return m, nil
		}
	case error:
		detail := NewDetail("Error")
		detail.SetSize(m.width, m.pageHeight())
		detail.SetContent(msg.Error())
		m.pages[len(m.pages)-1] = detail

		return m, detail.Init()
	}

	// Update the current page
	var cmd tea.Cmd

	currentPageIdx := len(m.pages) - 1
	m.pages[currentPageIdx], cmd = m.pages[currentPageIdx].Update(msg)

	return m, cmd
}

func (m *RootModel) View() string {
	if m.hidden {
		return ""
	}

	var embedView string
	if len(m.pages) > 0 {
		currentPage := m.pages[len(m.pages)-1]
		embedView = currentPage.View()
	} else {
		embedView = "No pages"
	}

	return embedView
}

func (m *RootModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	for _, page := range m.pages {
		page.SetSize(m.width, m.pageHeight())
	}
}

func (m *RootModel) pageHeight() int {
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
	container Container
}

func NewPushCmd(c Container) tea.Cmd {
	return func() tea.Msg {
		return pushMsg{c}
	}
}

func (m *RootModel) Push(page Container) tea.Cmd {
	page.SetSize(m.width, m.pageHeight())
	m.pages = append(m.pages, page)
	return page.Init()
}

func (m *RootModel) Pop() {
	if len(m.pages) > 0 {
		m.pages = m.pages[:len(m.pages)-1]
	}
}

func RootList(rootItems ...app.RootItem) *List {
	listItems := make([]ListItem, len(rootItems))
	for index, rootItem := range rootItems {
		with := make(app.ScriptInputs)
		for key, value := range rootItem.With {
			with[key] = app.ScriptInput{Value: value}
		}
		runMsg := RunScriptMsg{
			Extension: rootItem.Extension,
			Script:    rootItem.Script,
			With:      with,
		}
		listItems[index] = ListItem{
			Id:       strconv.Itoa(index),
			Title:    rootItem.Title,
			Subtitle: rootItem.Subtitle,
			Actions: []Action{
				{
					Title:    "Run Script",
					Shortcut: "enter",
					Cmd: func() tea.Msg {
						return runMsg
					},
				},
			},
		}
	}

	// Sort root items by title
	sort.SliceStable(listItems, func(i, j int) bool {
		return listItems[i].Title < listItems[j].Title
	})

	list := NewList("Sunbeam")
	list.SetItems(listItems)

	return list
}

func Draw(model *RootModel) (err error) {
	// Log to a file
	if env := os.Getenv("SUNBEAM_LOG_FILE"); env != "" {
		f, err := tea.LogToFile(env, "debug")
		if err != nil {
			log.Fatalf("could not open log file: %v", err)
		}
		defer f.Close()
	} else {
		tea.LogToFile("/dev/null", "")
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	if err != nil {
		return err
	}

	return nil
}
