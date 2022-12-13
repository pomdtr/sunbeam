package tui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/browser"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/skratchdot/open-golang/open"
)

type Config struct {
	Height      int
	AccentColor string

	RootItems []app.RootItem `yaml:"rootItems"`
}

type Container interface {
	Init() tea.Cmd
	Update(tea.Msg) (Container, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type RootModel struct {
	maxHeight     int
	width, height int
	exit          bool
	initCmd       tea.Cmd

	pages []Container

	hidden bool
}

func NewRootModel(initCmd tea.Cmd) *RootModel {
	return &RootModel{initCmd: initCmd}
}

func (m *RootModel) Init() tea.Cmd {
	return m.initCmd
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
		err := clipboard.WriteAll(msg.Text)
		if err != nil {
			return m, NewErrorCmd(err)
		}
		return m, tea.Quit
	case OpenUrlMsg:
		err := browser.OpenURL(msg.Url)
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

		params, formItems, err := extractScriptInput(script, msg.With)
		if err != nil {
			return m, NewErrorCmd(err)
		}

		if len(formItems) > 0 {
			form := NewForm(msg.Extension, formItems, func(values map[string]any) tea.Cmd {
				with := make(map[string]any)
				for key, value := range msg.With {
					with[key] = value
				}
				for key, value := range values {
					with[key] = value
				}

				msg.With = with

				return func() tea.Msg {
					return msg
				}
			})

			cmd := m.Push(form)
			return m, cmd
		}

		if script.IsPage() {
			runner := NewRunContainer(extension, script, params)
			cmd := m.Push(runner)
			return m, cmd
		}

		commandString, err := script.Cmd(params)

		switch script.Mode {
		case "command":
			return m, NewExecCmd(commandString, msg.Silent, msg.OnSuccess)
		case "snippet":
			return m, func() tea.Msg {
				command := exec.Command("sh", "-c", commandString)
				if err != nil {
					return err
				}
				command.Dir = extension.Dir()
				out, err := command.Output()
				if err != nil {
					return err
				}
				return CopyTextMsg{Text: string(out)}
			}
		case "quicklink":
			return m, func() tea.Msg {
				command := exec.Command("sh", "-c", commandString)
				if err != nil {
					return err
				}
				command.Dir = extension.Dir()
				out, err := command.Output()
				if err != nil {
					return err
				}
				return OpenUrlMsg{Url: string(out)}
			}
		default:
			return m, NewErrorCmd(fmt.Errorf("unknown mode: %s", script.Mode))
		}
	case ExecCommandMsg:
		command := exec.Command("sh", "-c", msg.Command)
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
				return pushMsg{container: detail}
			}

			return showMsg{
				cmd: msg.OnSuccessCmd,
			}
		})
	case pushMsg:
		m.hidden = false
		cmd := m.Push(msg.container)
		return m, cmd
	case popMsg:
		if len(m.pages) == 1 {
			m.hidden = true
			return m, tea.Quit
		} else {
			m.Pop()
			return m, nil
		}
	case showMsg:
		m.hidden = false
		return m, msg.cmd
	case error:
		detail := NewDetail("Error")
		detail.SetSize(m.width, m.pageHeight())
		detail.SetContent(msg.Error())

		currentIndex := len(m.pages) - 1
		m.pages[currentIndex] = detail
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
	if m.maxHeight > 0 {
		return utils.Min(m.maxHeight, m.height)
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
		runMsg := RunScriptMsg{
			Extension: rootItem.Extension,
			Script:    rootItem.Script,
			With:      rootItem.With,
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
