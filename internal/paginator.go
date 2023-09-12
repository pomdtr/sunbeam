package internal

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/pomdtr/sunbeam/pkg"
)

func PopPageCmd() tea.Msg {
	return PopPageMsg{}
}

type PopPageMsg struct{}

func NewPushPageCmd(page Page) tea.Cmd {
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

type SunbeamOptions struct {
	MaxHeight  int
	MaxWidth   int
	Border     bool
	FullScreen bool
	Margin     int
	NoColor    bool
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
	options       SunbeamOptions

	pages  []Page
	hidden bool
}

func CommandToPage(extension Extension, commandRef pkg.CommandRef) (Page, error) {
	command, ok := extension.Command(commandRef.Name)
	if !ok {
		return nil, fmt.Errorf("command %s not found", commandRef.Name)
	}

	switch command.Mode {
	case pkg.CommandModeFilter, pkg.CommandModeGenerator:
		return NewList(extension, command, commandRef.Params), nil
	case pkg.CommandModeForm:
		return NewDynamicForm(extension, command, commandRef.Params), nil
	case pkg.CommandModeDetail:
		return NewDetail(extension, command, commandRef.Params), nil
	default:
		return nil, fmt.Errorf("unsupported command mode: %s", command.Mode)
	}
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
	case error:
		if len(m.pages) > 0 {
			m.pages = m.pages[:len(m.pages)-1]
		}
		cmd := m.Push(NewErrorPage(msg))
		return m, cmd
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

	style := lipgloss.NewStyle().Margin(m.MarginVertical(lipgloss.Height(pageView)), m.MarginHorizontal(lipgloss.Width(pageView)))

	if m.options.Border {
		style = style.Border(lipgloss.RoundedBorder())
	}

	return style.Render(pageView)
}

func (m Paginator) MarginHorizontal(width int) int {
	if m.options.MaxWidth == 0 {
		return m.options.Margin
	}

	if m.options.MaxWidth > m.width {
		return m.options.Margin
	}

	return (m.width - width - 1) / 2
}

func (m Paginator) MarginVertical(height int) int {
	if !m.options.FullScreen {
		return m.options.Margin
	}

	if m.options.MaxHeight == 0 {
		return m.options.Margin
	}

	if m.options.MaxHeight > m.height {
		return m.options.Margin
	}

	return (m.height - height - 1) / 2
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

	if m.options.Margin > 0 {
		pageWidth -= 2 * m.options.Margin
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

	if m.options.Margin > 0 {
		height -= 2 * m.options.Margin
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

func Draw(page Page, options SunbeamOptions) error {
	if options.NoColor {
		lipgloss.SetColorProfile(termenv.Ascii)
	} else {
		lipgloss.SetColorProfile(termenv.NewOutput(os.Stdout).Profile)
	}

	paginator := NewPaginator(page, options)

	var p *tea.Program
	if options.FullScreen {
		p = tea.NewProgram(paginator, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(paginator, tea.WithOutput(os.Stderr))
	}

	_, err := p.Run()
	if err != nil {
		return err
	}

	return nil
}
