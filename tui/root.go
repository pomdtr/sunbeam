package tui

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/browser"
	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/utils"
)

type Config struct {
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
	MaxHeight     int
	Padding       int

	root  Page
	pages []Page

	hidden bool
}

func NewModel(root Page) *Model {
	return &Model{root: root}
}

func (m *Model) SetRoot(root Page) {
	m.root = root
}

func (m *Model) Init() tea.Cmd {
	return m.root.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			fmt.Sprintln("Escape")
		case tea.KeyCtrlC:
			m.hidden = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case OpenUrlMsg:
		err := browser.OpenURL(msg.Url)
		if err != nil {
			return m, NewErrorCmd(err)
		}

		m.hidden = true
		return m, tea.Quit
	case CopyTextMsg:
		err := clipboard.WriteAll(msg.Text)
		if err != nil {
			return m, NewErrorCmd(fmt.Errorf("failed to copy text to clipboard: %s", err))
		}

		m.hidden = true
		return m, tea.Quit
	case PushPageMsg:
		cmd := m.Push(msg.Page)
		return m, cmd
	case pushMsg:
		cmd := m.Push(msg.container)
		return m, cmd
	case popMsg:
		if len(m.pages) == 0 {
			m.hidden = true
			return m, tea.Quit
		} else {
			m.Pop()
			return m, nil
		}
	case error:
		detail := NewDetail("Error", msg.Error)
		detail.SetSize(m.pageWidth(), m.pageHeight())

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

func (m *Model) View() string {
	if m.hidden {
		return ""
	}

	var pageView string

	if len(m.pages) > 0 {
		currentPage := m.pages[len(m.pages)-1]
		pageView = currentPage.View()
	} else {
		pageView = m.root.View()
	}

	if m.MaxHeight > 0 {
		return lipgloss.NewStyle().Padding(m.Padding).Render(pageView)
	}

	return pageView

}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	m.root.SetSize(m.pageWidth(), m.pageHeight())
	for _, page := range m.pages {
		page.SetSize(m.pageWidth(), m.pageHeight())
	}
}

func (m *Model) pageWidth() int {
	return m.width - 2*m.Padding
}

func (m *Model) pageHeight() int {
	if m.MaxHeight > 0 {
		return utils.Min(m.MaxHeight, m.height) - 2*m.Padding
	}
	return m.height - 2*m.Padding
}

type popMsg struct{}

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
	page.SetSize(m.pageWidth(), m.pageHeight())
	m.pages = append(m.pages, page)
	return page.Init()
}

func (m *Model) Pop() {
	if len(m.pages) > 0 {
		m.pages = m.pages[:len(m.pages)-1]
	}
}

type RootList struct {
	extensions map[string]app.Extension
	rootItems  []app.RootItem
	list       *List
}

func NewRootList(extensions map[string]app.Extension, additionalRootItems ...app.RootItem) *RootList {
	rootItems := make([]app.RootItem, 0)
	for name, extension := range extensions {
		for _, rootItem := range extension.RootItems {
			rootItem.Extension = name
			rootItems = append(rootItems, rootItem)
		}
	}

	for _, rootItem := range additionalRootItems {
		if _, ok := extensions[rootItem.Extension]; !ok {
			continue
		}
		rootItems = append(rootItems, rootItem)
	}

	return &RootList{
		extensions: extensions,
		rootItems:  rootItems,
		list:       NewList("Sunbeam"),
	}
}

func (rl RootList) Init() tea.Cmd {
	return tea.Batch(rl.list.Init(), rl.RefreshItem)
}

func (rl RootList) RefreshItem() tea.Msg {
	listItems := make([]ListItem, 0)
	for _, rootItem := range rl.rootItems {
		rootItem := rootItem
		extension := rl.extensions[rootItem.Extension]
		command, ok := extension.GetCommand(rootItem.Command)
		if !ok {
			return nil
		}
		listItems = append(listItems, ListItem{
			Id:          fmt.Sprintf("%s-%s", rootItem.Extension, rootItem.Command),
			Title:       rootItem.Title,
			Subtitle:    extension.Title,
			Accessories: []string{rootItem.Extension},
			Actions: []Action{
				{
					Title:    "Run Command",
					Shortcut: "enter",
					Cmd: func() tea.Msg {
						return PushPageMsg{
							Page: NewCommandRunner(
								extension,
								command,
								rootItem.With,
							),
						}
					},
				},
			},
		})
	}

	return listItems
}

func (rl RootList) SetSize(width int, height int) {
	rl.list.SetSize(width, height)
}

func (rl RootList) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case []ListItem:
		rl.list.SetItems(msg)
	}

	var cmd tea.Cmd
	page, cmd := rl.list.Update(msg)

	rl.list = page.(*List)

	return rl, cmd
}

func (rl RootList) View() string {
	return rl.list.View()
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

	// Disable the background detection since we are only using ANSI colors
	lipgloss.SetHasDarkBackground(true)

	var p *tea.Program

	padding, ok := os.LookupEnv("SUNBEAM_PADDING")
	if ok || padding != "" {
		padding, err := strconv.Atoi(padding)
		if err != nil {
			return fmt.Errorf("could not parse SUNBEAM_PADDING: %w", err)
		}
		model.Padding = padding
	}

	height, ok := os.LookupEnv("SUNBEAM_HEIGHT")
	if !ok || height == "" {
		p = tea.NewProgram(model, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
	} else {
		height, err := strconv.Atoi(height)
		if err != nil {
			return fmt.Errorf("could not parse SUNBEAM_HEIGHT: %w", err)
		}

		model.MaxHeight = height
		p = tea.NewProgram(model, tea.WithOutput(os.Stderr))
	}

	_, err = p.Run()
	if err != nil {
		return err
	}

	return nil
}
