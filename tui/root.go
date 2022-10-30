package tui

import (
	"log"
	"os"
	"path"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
)

type Container interface {
	Init() tea.Cmd
	Update(tea.Msg) (Container, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type RootModel struct {
	width, height int
	pages         []Container
}

func NewRootModel(rootPage Container) *RootModel {
	return &RootModel{pages: []Container{rootPage}}
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
		for i, page := range m.pages {
			m.pages[i], _ = page.Update(msg)
		}
		return m, nil
	case PushMsg:
		m.Push(msg.Page)
		return m, msg.Page.Init()
	case popMsg:
		m.Pop()
		if len(m.pages) == 0 {
			return m, tea.Quit
		}
	}

	// Update the current page
	var cmd tea.Cmd
	currentPageIdx := len(m.pages) - 1
	m.pages[currentPageIdx], cmd = m.pages[currentPageIdx].Update(msg)
	return m, cmd
}

func (m *RootModel) View() string {
	if len(m.pages) == 0 {
		return ""
	}
	return m.pages[len(m.pages)-1].View()
}

func (m *RootModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	for _, page := range m.pages {
		page.SetSize(width, height)
	}
}

type PushMsg struct {
	Page Container
}

func NewPushCmd(page Container) tea.Cmd {
	return func() tea.Msg {
		return PushMsg{Page: page}
	}
}

type popMsg struct{}

func PopCmd() tea.Msg {
	return popMsg{}
}

func (m *RootModel) Push(page Container) {
	page.SetSize(m.width, m.height)
	m.pages = append(m.pages, page)
}

func (m *RootModel) Pop() {
	if len(m.pages) > 0 {
		m.pages = m.pages[:len(m.pages)-1]
	}
}

// Just a wrapper to convert the list to a container
type RootContainer struct {
	*List
}

func (c *RootContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	var cmd tea.Cmd
	c.List, cmd = c.List.Update(msg)
	return c, cmd
}

func Start() error {
	rootItems := make([]ListItem, 0)
	for _, manifest := range api.Sunbeam.Extensions {
		for _, item := range manifest.RootItems {
			if item.Subtitle == "" {
				item.Subtitle = manifest.Name
			}
			rootItems = append(rootItems, NewListItem(manifest.Name, item))
		}
	}

	list := NewList(false)
	list.SetItems(rootItems)
	rootContainer := RootContainer{List: list}
	return Draw(&rootContainer)
}

func Run(command api.SunbeamScript, params map[string]string) error {
	container := NewRunContainer(command, params)
	return Draw(container)
}

func Draw(container Container) (err error) {
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

	m := NewRootModel(container)
	p := tea.NewProgram(m, tea.WithAltScreen())
	err = p.Start()

	if err != nil {
		return err
	}
	return nil
}
