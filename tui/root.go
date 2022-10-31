package tui

import (
	"log"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/skratchdot/open-golang/open"
)

type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type RootModel struct {
	maxWidth, maxHeight int
	width, height       int

	pages []Page
}

func NewRootModel(width, height int, rootPage Page) *RootModel {
	return &RootModel{pages: []Page{rootPage}, maxWidth: width, maxHeight: height}
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
		m.Push(msg.Page)
		return m, msg.Page.Init()
	case popMsg:
		if len(m.pages) == 1 {
			return m, tea.Quit
		} else {
			m.Pop()
			return m, nil
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
		return "This should not happen, please report this bug"
	}

	var pageStyle lipgloss.Style
	pageStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true)
	currentPage := m.pages[len(m.pages)-1]
	return lipgloss.Place(m.width, m.height, lipgloss.Position(lipgloss.Center), lipgloss.Position(lipgloss.Center), pageStyle.Render(currentPage.View()))
}

func (m *RootModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	for _, page := range m.pages {
		m.setPageSize(page, width, height)
	}
}

func (m *RootModel) setPageSize(page Page, width, height int) {
	page.SetSize(utils.Min(m.maxWidth, width-2), utils.Min(m.maxHeight, height-2))
}

type PushMsg struct {
	Page Page
}

func NewPushCmd(page Page) tea.Cmd {
	return func() tea.Msg {
		return PushMsg{Page: page}
	}
}

type popMsg struct{}

func PopCmd() tea.Msg {
	return popMsg{}
}

func (m *RootModel) Push(page Page) {
	m.setPageSize(page, m.width, m.height)
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

func (c *RootContainer) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace:
			// Don't quit the app when the user press backspace
			if c.textInput.Value() == "" {
				return c, nil
			}
		}
	}

	c.List, cmd = c.List.Update(msg)
	return c, cmd
}

func Start(width, height int) error {
	rootItems := make([]ListItem, 0)
	for _, manifest := range api.Sunbeam.Extensions {
		for _, scriptAction := range manifest.RootActions {
			var accessoryTitle string
			action := NewAction(scriptAction)

			action.Shortcut = "enter"
			switch scriptAction.Type {
			case "run":
				action.Title = "Open Command"
				accessoryTitle = "Command"
			case "open-url":
				accessoryTitle = "Quicklink"
				action.Title = "Open Link"
			case "open-file":
				accessoryTitle = "Quicklink"
				action.Title = "Open File"
			case "copy":
				accessoryTitle = "Snippet"
				action.Title = "Copy Snippet"
			}

			rootItems = append(rootItems, ListItem{
				Title:       scriptAction.Title,
				Subtitle:    manifest.Title,
				Extension:   manifest.Name,
				Accessories: []string{accessoryTitle},
				Actions: []Action{
					action,
				},
			})
		}
	}

	list := NewList(false)
	list.SetItems(rootItems)
	rootContainer := RootContainer{List: list}

	m := NewRootModel(width, height, &rootContainer)
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
