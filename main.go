package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/commands"
	"github.com/pomdtr/sunbeam/containers"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	width  int
	height int
	pages  []containers.Container
}

func NewModel(pages ...containers.Container) model {
	return model{
		pages: pages,
	}
}

func (m *model) PushPage(page containers.Container) {
	m.pages = append(m.pages, page)
}

func (m *model) PopPage() {
	if len(m.pages) == 0 {
		return
	}
	m.pages = m.pages[:len(m.pages)-1]
}

func (m model) Init() tea.Cmd {
	if len(m.pages) == 0 {
		return nil
	}
	return m.pages[0].Init()
}

func (m *model) SetSize(width, height int) {
	m.width = width
	m.height = height
	for i := range m.pages {
		m.pages[i].SetSize(width, height)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case containers.PushMsg:
		page := msg.Container
		page.SetSize(m.width, m.height)
		m.PushPage(msg.Container)
		cmds = append(cmds, msg.Container.Init())
	case containers.PopMsg:
		m.PopPage()
		return m, nil

	case error:
		log.Printf("Error: %v", msg)
		return m, tea.Quit
	}

	if len(m.pages) == 0 {
		return m, nil
	}

	var cmd tea.Cmd
	currentPageIdx := len(m.pages) - 1
	m.pages[currentPageIdx], cmd = m.pages[currentPageIdx].Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)

}

func (m model) View() string {
	if len(m.pages) == 0 {
		return "No pages"
	}
	return m.pages[len(m.pages)-1].View()
}

func main() {
	// Log to a file
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatalf("could not open log file: %v", err)
	}
	defer f.Close()

	root := containers.NewRootContainer(commands.CommandDirs)
	m := NewModel(&root)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
