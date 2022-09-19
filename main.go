package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	pageFactory PageFactory
	pages       []tea.Model
}

type views struct {
	list   list.Model
	detail any
	form   any
}

func New(pageFactory PageFactory, pages ...tea.Model) model {
	return model{
		pageFactory: pageFactory,
		pages:       pages,
	}
}

func (m *model) PushPage(page tea.Model) {
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

type PushMsg struct {
	Script Script
	Args   []string
}

type PopMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEscape:
			if len(m.pages) == 1 {
				return m, tea.Quit
			}

			m.PopPage()
			return m, nil
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.pageFactory.SetSize(msg.Width-h, msg.Height-v)
		for i, page := range m.pages {
			page, _ := page.Update(msg)
			m.pages[i] = page
		}
		return m, nil

	case PushMsg:
		page := m.pageFactory.BuildPage(msg.Script)
		m.PushPage(page)
		return m, page.Init()

	case PopMsg:
		m.PopPage()

	case Script:
		page := m.pageFactory.BuildPage(msg)
		m.PushPage(page)
		return m, page.Init()

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
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatalf("could not open log file: %v", err)
	}
	defer f.Close()

	pageFactory := NewPageFactory()
	var commandDir = "/Users/a.lacoin/Developer/pomdtr/sunbeam/scripts"
	root := NewRootPage(commandDir)
	m := New(pageFactory, root)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
