package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func PopPageCmd() tea.Msg {
	return PopPageMsg{}
}

type PopPageMsg struct{}

func PushPageCmd(page Page) tea.Cmd {
	return func() tea.Msg {
		return PushPageMsg{
			Page: page,
		}
	}
}

type PushPageMsg struct {
	Page Page
}

type Page interface {
	Init() tea.Cmd
	SetSize(width, height int)
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
}

type WindowOptions struct {
	Height int
}

type ExitMsg struct{}

func ExitCmd() tea.Msg {
	return ExitMsg{}
}

type FocusMsg struct{}

func FocusCmd() tea.Msg {
	return FocusMsg{}
}

type Paginator struct {
	width, height int
	options       WindowOptions

	pages  []Page
	hidden bool
}

func NewPaginator(root Page, options WindowOptions) *Paginator {
	return &Paginator{pages: []Page{
		root,
	}, options: options}
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
		case tea.KeyCtrlC:
			m.hidden = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case PushPageMsg:
		cmd := m.Push(msg.Page)
		return m, cmd
	case PopPageMsg:
		if len(m.pages) > 1 {
			cmd := m.Pop()
			return m, cmd
		}

		m.hidden = true
		return m, tea.Quit
	case ExitMsg:
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

	if len(m.pages) > 0 {
		currentPage := m.pages[len(m.pages)-1]
		if m.options.Height > 0 {
			return lipgloss.NewStyle().PaddingTop(1).Render(currentPage.View())
		}
		return currentPage.View()
	}

	return ""
}

func (m *Paginator) SetSize(width, height int) {
	m.width = width
	m.height = height

	for _, page := range m.pages {
		page.SetSize(m.pageWidth(), m.pageHeight())
	}
}

func (m *Paginator) pageWidth() int {
	return m.width
}

func (m *Paginator) pageHeight() int {
	if m.options.Height == 0 {
		return m.height
	}

	height := min(m.height, m.options.Height)
	if height > 0 {
		return height - 1 // margin top
	}

	return height
}

func (m *Paginator) Push(page Page) tea.Cmd {
	page.SetSize(m.pageWidth(), m.pageHeight())
	m.pages = append(m.pages, page)
	return page.Init()
}

func (m *Paginator) Pop() tea.Cmd {
	if len(m.pages) > 0 {
		m.pages = m.pages[:len(m.pages)-1]
	}

	return func() tea.Msg {
		return FocusMsg{}
	}
}

func Draw(page Page, options WindowOptions) error {
	lipgloss.SetColorProfile(termenv.NewOutput(os.Stderr).Profile)
	paginator := NewPaginator(page, options)

	var p *tea.Program
	if options.Height > 0 {
		p = tea.NewProgram(paginator, tea.WithOutput(os.Stderr))
	} else {
		p = tea.NewProgram(paginator, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
	}

	_, err := p.Run()
	if err != nil {
		return err
	}

	return nil
}
