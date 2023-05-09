package internal

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type PopPageMsg struct{}

type PushPageMsg struct {
	Page Page
}

type Page interface {
	Init() tea.Cmd
	Focus() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type SunbeamOptions struct {
	MaxHeight  int
	MaxWidth   int
	Border     bool
	FullScreen bool
}

type ExitMsg struct {
	Cmd *exec.Cmd
}

type Paginator struct {
	width, height int
	options       SunbeamOptions
	OutputCmd     *exec.Cmd

	pages  []Page
	hidden bool
}

func NewPaginator(root Page, options SunbeamOptions) *Paginator {
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
		m.OutputCmd = msg.Cmd
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

	style := lipgloss.NewStyle()
	if m.options.Border {
		style = style.Border(lipgloss.RoundedBorder())
	}

	marginHorizontal := 0
	if m.options.MaxWidth != 0 && m.options.MaxWidth < m.width {
		marginHorizontal = (m.width - lipgloss.Width(pageView) - 1) / 2
	}

	marginVertical := 0
	if m.options.MaxHeight != 0 && m.options.MaxHeight < m.height && m.options.FullScreen {
		marginVertical = (m.height - lipgloss.Height(pageView) - 1) / 2
	}

	return style.Margin(marginVertical, marginHorizontal).Render(pageView)
}

func (m *Paginator) SetSize(width, height int) {
	m.width = width
	m.height = height

	for _, page := range m.pages {
		page.SetSize(m.pageWidth(), m.pageHeight())
	}
}

func (m *Paginator) pageWidth() int {
	pageWidth := m.width

	if m.options.MaxWidth > 0 && m.options.MaxWidth < pageWidth {
		pageWidth = m.options.MaxWidth
	}

	if m.options.Border {
		pageWidth -= 2
	}

	return pageWidth
}

func (m *Paginator) pageHeight() int {
	height := m.height

	if m.options.MaxHeight > 0 && m.options.MaxHeight < height {
		height = m.options.MaxHeight
	}

	if m.options.Border {
		height -= 2
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

	page := m.pages[len(m.pages)-1]
	return page.Focus()
}

func Draw(page Page, options SunbeamOptions) error {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		lipgloss.SetColorProfile(termenv.Ascii)
	} else {
		lipgloss.SetColorProfile(termenv.NewOutput(os.Stderr).Profile)
	}

	paginator := NewPaginator(page, options)

	var p *tea.Program
	if options.FullScreen {
		p = tea.NewProgram(paginator, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
	} else {
		p = tea.NewProgram(paginator, tea.WithOutput(os.Stderr))
	}

	m, err := p.Run()
	if err != nil {
		return err
	}

	paginator, ok := m.(*Paginator)
	if !ok {
		return fmt.Errorf("could not cast model to paginator")
	}

	cmd := paginator.OutputCmd
	if cmd == nil {
		return nil
	}

	if cmd.Stdin == nil {
		cmd.Stdin = os.Stdin
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}
