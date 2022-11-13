package tui

import (
	"fmt"
	"log"
	"os"
	"path"
	"sort"

	"github.com/adrg/xdg"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/skratchdot/open-golang/open"
)

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
	case CopyMsg:
		err := clipboard.WriteAll(msg.Content)
		if err != nil {
			m.exitMsg = fmt.Sprintf("Failed to copy to clipboard: %v", err)
			return m, NewErrorCmd(err)
		}
		m.exitMsg = "Copied to clipboard"
		return m, tea.Quit
	case OpenUrlMsg, OpenFileMessage:
		var err error
		var target, application string
		switch msg := msg.(type) {
		case OpenUrlMsg:
			target = msg.Url
			application = msg.Application
		case OpenFileMessage:
			target = msg.Path
			application = msg.Application
		}
		if application != "" {
			err = open.RunWith(target, application)
			m.exitMsg = fmt.Sprintf("Opened %s with %s", target, application)
		} else {
			err = open.Run(target)
			m.exitMsg = fmt.Sprintf("Opened %s", target)
		}
		if err != nil {
			m.exitMsg = fmt.Sprintf("Failed to open %s: %v", target, err)
			return m, NewErrorCmd(err)
		}
		return m, tea.Quit
	case PushMsg:
		manifest, ok := api.Sunbeam.Extensions[msg.Extension]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("extension %s not found", msg.Extension))
		}
		page, ok := manifest.Pages[msg.Page]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("page %s not found", msg.Page))
		}

		missing := page.CheckMissingParams(msg.Params)
		if len(missing) > 0 {
			var err error
			items := make([]FormItem, len(missing))
			for i, param := range missing {
				items[i], err = NewFormItem(param)
				if err != nil {
					return m, NewErrorCmd(err)
				}
			}
			form := NewForm(page.Title, items, func(values map[string]any) tea.Cmd {
				params := make(map[string]any)
				for k, v := range msg.Params {
					params[k] = v
				}
				for k, v := range values {
					params[k] = v
				}
				return func() tea.Msg {
					return PushMsg{
						Extension: msg.Extension,
						Page:      msg.Page,
						Params:    params,
					}
				}
			})
			m.Push(form)
			return m, form.Init()
		}

		if page.Title == "" {
			page.Title = msg.Page
		}
		runner := NewRunContainer(manifest, page, msg.Params)
		m.Push(runner)
		return m, runner.Init()
	case RunMsg:
		manifest, ok := api.Sunbeam.Extensions[msg.Extension]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("extension %s not found", msg.Extension))
		}
		script := manifest.Scripts[msg.Script]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("script %s does exists in extension %s", msg.Script, msg.Extension))
		}

		missing := script.CheckMissingParams(msg.With)
		if len(missing) > 0 {
			items := make([]FormItem, len(missing))
			var err error
			for i, param := range missing {

				items[i], err = NewFormItem(param)
				if err != nil {
					return m, NewErrorCmd(err)
				}
			}
			if script.Title == "" {
				script.Title = msg.Script
			}
			form := NewForm(script.Title, items, func(values map[string]any) tea.Cmd {
				params := make(map[string]any)
				for k, v := range msg.With {
					params[k] = v
				}
				for k, v := range values {
					params[k] = v
				}
				return func() tea.Msg {
					return RunMsg{
						Extension: msg.Extension,
						Script:    msg.Script,
						With:      params,
					}
				}
			})
			m.Push(form)
			return m, form.Init()
		}

		// Run the script
		output, err := script.Run(manifest.Dir(), msg.With)
		if err != nil {
			return m, NewErrorCmd(err)
		}

		return m, func() tea.Msg {
			switch msg.OnSuccess {
			case "copy-content":
				return CopyMsg{
					Content: output,
				}
			case "open-url":
				return OpenUrlMsg{
					Url: output,
				}
			case "open-file":
				return OpenFileMessage{
					Path: output,
				}
			case "reload-page":
				return ReloadMsg{}
			default:
				m.exitMsg = output
				return tea.Quit()
			}
		}
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

func Start(width, height int) error {
	entrypoints := make([]ListItem, 0)
	for _, manifest := range api.Sunbeam.Extensions {
		for _, rootItem := range manifest.Entrypoints {
			title := rootItem.Title
			rootItem.Title = "Open Command"
			rootItem.Extension = manifest.Name
			rootItem.Shortcut = "enter"
			for key, param := range rootItem.With {
				param, ok := param.(string)
				if !ok {
					continue
				}

				param, err := utils.RenderString(param, nil)
				if err != nil {
					log.Printf("failed to render param %s: %v", param, err)
					continue
				}

				rootItem.With[key] = param
			}
			entrypoints = append(entrypoints, ListItem{
				Title:    title,
				Subtitle: manifest.Title,
				Actions: []Action{
					NewAction(rootItem),
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

	m := NewRootModel(width, height, list)
	return Draw(m)
}

// func Run(command api.SunbeamScript, params map[string]string) error {
// 	container := NewRunContainer(command, params)
// 	return Draw(container)
// }

func Draw(model tea.Model) (err error) {
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

	p := tea.NewProgram(model, tea.WithAltScreen())
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
