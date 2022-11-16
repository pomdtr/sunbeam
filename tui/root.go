package tui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/alessio/shellescape"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/browser"
	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/skratchdot/open-golang/open"
)

type SunbeamOptions struct {
	MaxWidth  int
	MaxHeight int
	Theme     string
	Accent    string
}

type Container interface {
	Init() tea.Cmd
	Update(tea.Msg) (Container, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type RootModel struct {
	maxWidth, maxHeight int
	width, height       int

	pages   []Container
	exitMsg string
	exitCmd *exec.Cmd
}

func NewRootModel(width, height int, rootPage Container) *RootModel {
	return &RootModel{pages: []Container{rootPage}, maxWidth: width, maxHeight: height}
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
		return m, tea.Quit
	case OpenUrlMsg:
		err := browser.OpenURL(msg.Url)
		if err != nil {
			m.exitMsg = fmt.Sprintf("Failed to open url: %v", err)
		}
		m.exitMsg = fmt.Sprintf("Opened %s in browser.", msg.Url)
		return m, tea.Quit
	case OpenFileMessage:
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
		return m, tea.Quit
	case RunScriptMsg:
		manifest, ok := api.Sunbeam.Extensions[msg.Extension]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("extension %s not found", msg.Extension))
		}
		script, ok := manifest.Scripts[msg.Script]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("page %s not found", msg.Script))
		}

		if script.Page.Title == "" {
			script.Page.Title = msg.Script
		}
		runner := NewRunContainer(manifest, script, msg.With)
		m.Push(runner)
		return m, runner.Init()
	case ExecMsg:
		// Run the script
		if msg.OnSuccess == "reload-page" {
			err := msg.Command.Run()
			if err != nil {
				return m, NewErrorCmd(err)
			}
			return m, func() tea.Msg {
				return ReloadPageMsg{}
			}
		}

		m.exitCmd = msg.Command
		return m, tea.Quit
	case popMsg:
		if len(m.pages) == 1 {
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
	var embedView string
	if len(m.pages) == 0 {
		return "This should not happen, please report this bug"
	}

	currentPage := m.pages[len(m.pages)-1]
	embedView = currentPage.View()

	pageStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderBackground(theme.Bg()).BorderForeground(theme.Fg())
	return lipgloss.Place(m.width, m.height, lipgloss.Position(lipgloss.Center), lipgloss.Position(lipgloss.Center), pageStyle.Render(embedView), lipgloss.WithWhitespaceBackground(theme.Bg()))
}

func (m *RootModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	for _, page := range m.pages {
		page.SetSize(m.pageWidth(), m.pageHeight())
	}
}

func (m *RootModel) pageWidth() int {
	return utils.Min(m.maxWidth, m.width-2)
}

func (m *RootModel) pageHeight() int {
	return utils.Min(m.maxHeight, m.height-2)
}

type popMsg struct{}

func PopCmd() tea.Msg {
	return popMsg{}
}

func (m *RootModel) Push(page Container) {
	page.SetSize(m.pageWidth(), m.pageHeight())
	m.pages = append(m.pages, page)
}

func (m *RootModel) Pop() {
	if len(m.pages) > 0 {
		m.pages = m.pages[:len(m.pages)-1]
	}
}

func ScriptRunCmd(extension string, script string, params map[string]any) tea.Cmd {
	return func() tea.Msg {
		args := make([]string, 0)
		args = append(args, "sunbeam", "run", extension, script)
		for param, value := range params {
			switch value := value.(type) {
			case string:
				value = shellescape.Quote(value)
				args = append(args, fmt.Sprintf("--with %s=%s", param, value))
			case bool:
				args = append(args, fmt.Sprintf("--with %s=%t", param, value))
			}
		}
		return CopyTextMsg{
			Text: strings.Join(args, " "),
		}
	}
}

func RootList(manifests ...api.Manifest) Container {
	entrypoints := make([]ListItem, 0)
	for _, manifest := range manifests {
		for i, entrypoint := range manifest.Entrypoints {
			entrypoint := entrypoint
			entrypoints = append(entrypoints, ListItem{
				id:       strconv.Itoa(i),
				Title:    entrypoint.Title,
				Subtitle: manifest.Title,
				Actions: []Action{
					{
						Title:    "Run Script",
						Shortcut: "enter",
						Cmd: func() tea.Msg {
							return RunScriptMsg{
								Extension: manifest.Name,
								Script:    entrypoint.Script,
								With:      entrypoint.With,
							}
						},
					},
					{
						Title:    "Copy Shortcut",
						Shortcut: "ctrl+y",
						Cmd:      ScriptRunCmd(manifest.Name, entrypoint.Script, entrypoint.With),
					},
				},
			})
		}
	}

	// Sort entrypoints by title
	sort.SliceStable(entrypoints, func(i, j int) bool {
		return entrypoints[i].Title < entrypoints[j].Title
	})

	list := NewList("Sunbeam")
	list.SetItems(entrypoints)

	return list
}

func Draw(container Container, options SunbeamOptions) (err error) {
	var logFile string
	// Log to a file
	if env := os.Getenv("SUNBEAM_LOG_FILE"); env != "" {
		logFile = env
	} else {
		if _, err := os.Stat(xdg.StateHome); os.IsNotExist(err) {
			err = os.MkdirAll(xdg.StateHome, 0755)
			if err != nil {
				log.Fatalln(err)
			}
		}
		logFile = path.Join(xdg.StateHome, "api.log")
	}
	f, err := tea.LogToFile(logFile, "debug")
	if err != nil {
		log.Fatalf("could not open log file: %v", err)
	}
	defer f.Close()

	model := NewRootModel(options.MaxWidth, options.MaxHeight, container)
	p := tea.NewProgram(model, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return err
	}

	root, _ := m.(*RootModel)

	if root.exitMsg != "" {
		fmt.Println(root.exitMsg)
	}

	if exitCmd := root.exitCmd; exitCmd != nil {
		root.exitCmd.Stderr = os.Stderr
		root.exitCmd.Stdout = os.Stdout
		err := root.exitCmd.Run()
		if err != nil {
			return err
		}
	}

	return nil

}
