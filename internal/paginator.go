package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/browser"
	"github.com/muesli/termenv"
	"github.com/pomdtr/sunbeam/pkg"
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
	Margin     int
	NoColor    bool
}

type ExitMsg struct{}

type Paginator struct {
	width, height int
	options       SunbeamOptions
	Cancelled     bool
	err           error

	pages  []Page
	hidden bool
}

func CommandToPage(extensions Extensions, extensionName string, commandName string, params map[string]any) (Page, error) {
	extension, err := extensions.Get(extensionName)
	if err != nil {
		return nil, err
	}

	command, ok := extension.Command(commandName)
	if !ok {
		return nil, fmt.Errorf("command %s not found", commandName)
	}

	runner := func(action pkg.Action) tea.Cmd {
		return func() tea.Msg {
			switch action.Type {
			case pkg.ActionTypeCopy:
				if err := clipboard.WriteAll(action.Text); err != nil {
					return err
				}

				if action.Exit {
					return ExitMsg{}
				}

				return nil
			case pkg.ActionTypeOpen:
				if err := browser.OpenURL(action.Url); err != nil {
					return err
				}

				if action.Exit {
					return ExitMsg{}
				}

				return nil
			case pkg.ActionTypeRun:
				if command.Mode == pkg.CommandModeSilent {
					_, err := extension.Run(action.Command, pkg.CommandInput{
						Params: action.Params,
					})
					if err != nil {
						return err
					}

					if action.Exit {
						return ExitMsg{}
					}

					return nil
				}

				page, err := CommandToPage(extensions, extensionName, action.Command, action.Params)
				if err != nil {
					return err
				}

				return PushPageMsg{Page: page}
			}

			return nil
		}
	}

	switch command.Mode {
	case pkg.CommandModeFilter, pkg.CommandModeGenerator:
		return NewList(command.Title, func() (pkg.List, error) {
			var list pkg.List

			res, err := extension.Run(command.Name, pkg.CommandInput{
				Params: params,
			})
			if err != nil {
				return list, err
			}

			if err := json.Unmarshal(res, &list); err != nil {
				return list, err
			}

			return list, nil
		}, runner), nil
	case pkg.CommandModeDetail:
		return NewDetail(command.Title, func() (pkg.Detail, error) {
			var detail pkg.Detail

			res, err := extension.Run(command.Name, pkg.CommandInput{
				Params: params,
			})
			if err != nil {
				return detail, err
			}

			if err := json.Unmarshal(res, &detail); err != nil {
				return detail, err
			}

			return detail, nil
		}, runner), nil
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
			m.Cancelled = true
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

		m.Cancelled = true
		m.hidden = true
		return m, tea.Quit
	case ExitMsg:
		m.hidden = true
		return m, tea.Quit
	case error:
		m.err = msg
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

	if m.err != nil {
		return fmt.Sprintf("Error: %s", m.err)
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

	page := m.pages[len(m.pages)-1]
	return page.Focus()
}

var ErrInterrupted = errors.New("interrupted")

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

	m, err := p.Run()
	if err != nil {
		return err
	}

	paginator, ok := m.(*Paginator)
	if !ok {
		return fmt.Errorf("could not cast model to paginator")
	}

	if paginator.Cancelled {
		return ErrInterrupted
	}

	return nil
}
