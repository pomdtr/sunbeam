package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/utils"
)

type PopPageMsg struct{}

type PushPageMsg struct {
	runner Page
}

type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type SunbeamOptions struct {
	MaxHeight int
	Padding   int
}

type Paginator struct {
	width, height int
	options       SunbeamOptions

	pages []Page

	hidden bool
}

func NewPaginator(root Page) *Paginator {
	return &Paginator{pages: []Page{
		root,
	}, options: SunbeamOptions{
		MaxHeight: utils.LookupInt("SUNBEAM_HEIGHT", 0),
		Padding:   utils.LookupInt("SUNBEAM_PADDING", 0),
	}}
}

func (m *Paginator) Init() tea.Cmd {
	if len(m.pages) == 0 {
		return nil
	}

	return m.pages[0].Init()
}

func (m *Paginator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case PushPageMsg:
		cmd := m.Push(msg.runner)
		return m, cmd
	case PopPageMsg:
		if len(m.pages) > 1 {
			m.Pop()
			return m, nil
		}

		m.hidden = true
		return m, tea.Quit
	}

	// Update the current page
	var cmd tea.Cmd

	if len(m.pages) > 0 {
		currentPageIdx := len(m.pages) - 1
		m.pages[currentPageIdx], cmd = m.pages[currentPageIdx].Update(msg)
	} else {
		return m, nil
	}

	return m, cmd
}

func (m *Paginator) View() string {
	if m.hidden {
		return ""
	}

	var pageView string
	if len(m.pages) > 0 {
		currentPage := m.pages[len(m.pages)-1]
		pageView = currentPage.View()
	}

	return lipgloss.NewStyle().Padding(m.options.Padding).Render(pageView)
}

func (m *Paginator) SetSize(width, height int) {
	m.width = width
	m.height = height

	for _, page := range m.pages {
		page.SetSize(m.pageWidth(), m.pageHeight())
	}
}

func (m *Paginator) pageWidth() int {
	return m.width - 2*m.options.Padding
}

func (m *Paginator) pageHeight() int {
	if m.options.MaxHeight > 0 {
		return utils.Min(m.options.MaxHeight, m.height) - 2*m.options.Padding
	}
	return m.height - 2*m.options.Padding
}

func (m *Paginator) Push(page Page) tea.Cmd {
	page.SetSize(m.pageWidth(), m.pageHeight())
	m.pages = append(m.pages, page)
	return page.Init()
}

func (m *Paginator) Pop() {
	if len(m.pages) > 0 {
		m.pages = m.pages[:len(m.pages)-1]
	}
}

func (m *Paginator) Draw() (err error) {
	// Background detection before we start the program
	lipgloss.SetHasDarkBackground(lipgloss.HasDarkBackground())

	var p *tea.Program
	if m.options.MaxHeight == 0 {
		p = tea.NewProgram(m, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(m)
	}

	_, err = p.Run()
	return err
}
