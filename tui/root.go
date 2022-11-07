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

	pages []Container
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
			return m, NewErrorCmd(err)
		}
		return m, tea.Quit
	case OpenMsg:
		var err error
		if msg.Application != "" {
			err = open.RunWith(msg.Url, msg.Application)
		} else {
			err = open.Run(msg.Url)
		}
		if err != nil {
			return m, NewErrorCmd(err)
		}
		return m, tea.Quit
	case PushMsg:
		manifest, ok := api.Sunbeam.Extensions[msg.Extension]
		if !ok {
			return m, NewErrorCmd(fmt.Errorf("extension %s not found", msg.Extension))
		}
		page := NewRunContainer(manifest, msg.Page, msg.Params)
		m.Push(page)
		return m, page.Init()
	case popMsg:
		if len(m.pages) == 1 {
			return m, tea.Quit
		} else {
			m.Pop()
			return m, nil
		}
	case error:
		detail := NewDetail("Error")
		detail.SetContent(msg.Error())
		detail.SetSize(m.pageWidth(), m.pageHeight())

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
	if len(m.pages) == 0 {
		return "This should not happen, please report this bug"
	}

	currentPage := m.pages[len(m.pages)-1]
	view := currentPage.View()

	var pageStyle lipgloss.Style
	pageStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true)
	return lipgloss.Place(m.width, m.height, lipgloss.Position(lipgloss.Center), lipgloss.Position(lipgloss.Center), pageStyle.Render(view))
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
		for _, rootItem := range manifest.RootItems {
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

	// Necessary to cache the style
	lipgloss.HasDarkBackground()

	p := tea.NewProgram(model, tea.WithAltScreen())
	err = p.Start()

	if err != nil {
		return err
	}
	return nil
}
