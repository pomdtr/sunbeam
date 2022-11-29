package tui

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/browser"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/skratchdot/open-golang/open"
)

type Config struct {
	Height int
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

	quitting bool
	exitMsg  string
	exitCmd  *exec.Cmd
}

func NewRootModel(height int) *RootModel {
	return &RootModel{maxHeight: height}
}

func (m *RootModel) Init() tea.Cmd {
	if len(m.pages) > 0 {
		return m.pages[0].Init()
	}
	return nil
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
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
		m.quitting = true
		return m, tea.Quit
	case OpenUrlMsg:
		err := browser.OpenURL(msg.Url)
		if err != nil {
			m.exitMsg = fmt.Sprintf("Failed to open url: %v", err)
		}
		m.exitMsg = fmt.Sprintf("Opened %s in browser.", msg.Url)
		m.quitting = true
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
		m.quitting = true
		return m, tea.Quit
	case RunScriptMsg:
		extension, ok := app.Sunbeam.Extensions[msg.Extension]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("extension %s not found", msg.Extension))
		}
		script, ok := extension.Scripts[msg.Script]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("script %s not found", msg.Script))
		}

		missing := script.CheckMissingParams(msg.With)
		if len(missing) > 0 {
			return m, NewErrorCmd(fmt.Errorf("missing params: %s", strings.Join(missing, ", ")))
		}

		inputs := make(map[string]app.FormInput)
		for _, arg := range msg.With {
			if value, ok := arg.Value.(map[string]any); ok {
				bytes, err := json.Marshal(value)
				if err != nil {
					return m, NewErrorCmd(err)
				}
				var input app.FormInput
				err = json.Unmarshal(bytes, &input)
				if err != nil {
					return m, NewErrorCmd(err)
				}
				inputs[arg.Param] = input
			}
		}

		if len(inputs) > 0 {
			items := make([]FormItem, 0, len(inputs))
			for key, input := range inputs {
				formItem, err := NewFormItem(key, input)
				if err != nil {
					return m, NewErrorCmd(err)
				}
				items = append(items, formItem)
			}

			form := NewForm(msg.Extension, items, func(values map[string]any) tea.Cmd {
				for key, value := range values {
					for i, arg := range msg.With {
						if arg.Param == key {
							msg.With[i].Value = value
						}
					}
				}
				return tea.Sequence(PopCmd, func() tea.Msg {
					return msg
				})
			})

			cmd := m.Push(form)
			return m, cmd
		}

		if msg.OnSuccess != "" {
			script.OnSuccess = msg.OnSuccess
		}

		params := make(map[string]any, len(msg.With))
		for _, arg := range msg.With {
			params[arg.Param] = arg.Value
		}
		if script.OnSuccess == "push-page" {

			runner := NewRunContainer(extension, script, params)
			cmd := m.Push(runner)
			return m, cmd
		}

		command, err := script.Cmd(app.CommandInput{
			Params: params,
		})
		command.Dir = extension.Dir()
		if err != nil {
			return m, NewErrorCmd(err)
		}
		switch script.OnSuccess {
		case "copy-to-clipboard":
			return m, func() tea.Msg {
				out, err := command.Output()
				if err != nil {
					return err
				}
				return CopyTextMsg{Text: string(out)}
			}
		case "open-in-browser":
			return m, func() tea.Msg {
				out, err := command.Output()
				if err != nil {
					return err
				}
				return OpenUrlMsg{Url: string(out)}
			}
		case "reload-page":
			return m, func() tea.Msg {
				err := command.Run()
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					return NewErrorCmd(fmt.Errorf("script exited with code %d: %s", exitErr.ExitCode(), exitErr.Error()))
				} else if err != nil {
					return NewErrorCmd(err)
				}

				return ReloadPageMsg{}
			}
		case "":
			return m, NewExecCmd(command)
		default:
			return m, NewErrorCmd(fmt.Errorf("unknown onSuccess: %s", script.OnSuccess))
		}

	case ExecCommandMsg:
		m.quitting = true
		m.exitCmd = msg.Command
		return m, tea.Quit

	case popMsg:
		if len(m.pages) == 1 {
			m.quitting = true
			return m, tea.Quit
		} else {
			m.Pop()
			return m, nil
		}
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
	if m.quitting {
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

func PopCmd() tea.Msg {
	return popMsg{}
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

func ScriptCommand(extension string, entrypoint app.Entrypoint) string {
	args := make([]string, 0)
	args = append(args, "sunbeam", extension, entrypoint.Script)
	for _, arg := range entrypoint.With {
		switch value := arg.Value.(type) {
		case string:
			args = append(args, fmt.Sprintf("--%s=%s", arg.Param, value))
		case bool:
			args = append(args, fmt.Sprintf("--%s=%t", arg.Param, value))
		case map[string]interface{}:
			v, _ := json.Marshal(value)
			args = append(args, fmt.Sprintf("--%s=%s", arg.Param, shellescape.Quote(string(v))))
		}
	}
	return strings.Join(args, " ")
}

func RootList(extensions ...app.Extension) Container {
	rootItems := make([]ListItem, 0)
	for _, extension := range extensions {
		extension := extension
		for _, entrypoint := range extension.RootItems {
			runMsg := RunScriptMsg{
				Extension: extension.Id,
				Script:    entrypoint.Script,
				With:      entrypoint.With,
			}
			command := ScriptCommand(extension.Id, entrypoint)
			rootItems = append(rootItems, ListItem{
				Id:       command,
				Title:    entrypoint.Title,
				Subtitle: extension.Title,
				Actions: []Action{
					{
						Title:    "Run Script",
						Shortcut: "enter",
						Cmd: func() tea.Msg {
							return runMsg
						},
					}, {
						Title:    "Open Extension Folder",
						Shortcut: "ctrl+o",
						Cmd:      func() tea.Msg { return OpenPathMsg{Path: extension.Dir()} },
					},
					{
						Title:    "Copy Full Command",
						Shortcut: "ctrl+y",
						Cmd:      func() tea.Msg { return CopyTextMsg{Text: command} },
					},
				},
			})
		}
	}

	// Sort root items by title
	sort.SliceStable(rootItems, func(i, j int) bool {
		return rootItems[i].Title < rootItems[j].Title
	})

	list := NewList("Sunbeam")
	list.SetItems(rootItems)

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

	if exitCmd := root.exitCmd; exitCmd != nil {
		root.exitCmd.Stdin = os.Stdin
		root.exitCmd.Stdout = os.Stdout
		root.exitCmd.Stderr = os.Stderr
		err := root.exitCmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
