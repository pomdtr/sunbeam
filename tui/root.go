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
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/browser"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/skratchdot/open-golang/open"
)

type Config struct {
	Height    int
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

	pages []Container

	hidden  bool
	exitMsg string
}

func NewRootModel(height int) *RootModel {
	return &RootModel{maxHeight: height}
}

func (m *RootModel) Init() tea.Cmd {
	if len(m.pages) == 0 {
		return nil
	}
	return m.pages[0].Init()
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case CopyTextMsg:
		err := clipboard.WriteAll(msg.Text)
		if err != nil {
			m.exitMsg = fmt.Sprintf("Failed to copy to clipboard: %v", err)
			return m, NewErrorCmd(err)
		}
		m.exitMsg = "Copied to clipboard"
		m.hidden = true
		return m, tea.Quit
	case OpenUrlMsg:
		err := browser.OpenURL(msg.Url)
		if err != nil {
			m.exitMsg = fmt.Sprintf("Failed to open url: %v", err)
		}
		m.exitMsg = fmt.Sprintf("Opened %s in browser.", msg.Url)
		m.hidden = true
		return m, tea.Quit
	case OpenPathMsg:
		var err error

		if msg.Application != "" {
			err = open.RunWith(msg.Path, msg.Application)
			m.exitMsg = fmt.Sprintf("Opened %s with %s", msg.Path, msg.Application)
		} else {
			err = open.Run(msg.Path)
			m.exitMsg = fmt.Sprintf("Opened %s", msg.Path)
		}
		if err != nil {
			m.exitMsg = fmt.Sprintf("Failed to open %s: %v", msg.Path, err)
			return m, NewErrorCmd(err)
		}
		m.hidden = true
		return m, tea.Quit
	case RunScriptMsg:
		extension, ok := app.Sunbeam.Extensions[msg.Extension]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("extension %s not found", msg.Extension))
		}

		if len(extension.Requirements) > 0 {
			for _, requirement := range extension.Requirements {
				if !requirement.Check() {
					return m, NewErrorCmd(fmt.Errorf("requirement %s not met.\nHomepage: %s", requirement.Which, requirement.HomePage))
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
				msg := msg

				with := make(map[string]app.ScriptInput)
				for key, param := range msg.With {
					if value, ok := values[key]; ok {
						param.Value = value
					}
					with[key] = param
				}
				msg.With = with

				return tea.Sequence(PopCmd, func() tea.Msg {
					return msg
				})
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
		case "clipboard":
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
		case "browser":
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
		detail.SetSize(m.pageWidth(), m.pageHeight())
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

	pageStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.Fg())

	if m.maxHeight > 0 {
		return pageStyle.Render(embedView)
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Position(lipgloss.Center), lipgloss.Position(lipgloss.Center), pageStyle.Render(embedView))
}

func (m *RootModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	for _, page := range m.pages {
		page.SetSize(m.pageWidth(), m.pageHeight())
	}
}

func (m *RootModel) pageWidth() int {
	return utils.Max(0, m.width-2)
}

func (m *RootModel) pageHeight() int {
	var height int
	if m.maxHeight > 0 {
		height = utils.Min(m.maxHeight, m.height)
	} else {
		height = m.height
	}

	return utils.Max(0, height-2)
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
	page.SetSize(m.pageWidth(), m.pageHeight())
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

func Draw(container Container, config Config) (err error) {
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

	programOptions := make([]tea.ProgramOption, 0)
	if config.Height == 0 {
		programOptions = append(programOptions, tea.WithAltScreen())
	}

	model := NewRootModel(config.Height)
	model.Push(container)
	p := tea.NewProgram(model, programOptions...)
	m, err := p.Run()
	if err != nil {
		return err
	}

	root, _ := m.(*RootModel)

	if root.exitMsg != "" {
		fmt.Println(root.exitMsg)
	}

	return nil
}
