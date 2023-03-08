package tui

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/browser"
	"github.com/pomdtr/sunbeam/utils"
)

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

type Model struct {
	width, height int
	options       SunbeamOptions
	exitCmd       *exec.Cmd

	root  Page
	pages []Page

	hidden bool
}

func NewModel(root Page, options SunbeamOptions) *Model {
	return &Model{root: root, options: options}
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
	case OpenMsg:
		err := browser.OpenURL(msg.Target)
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
	case exit:
		m.hidden = true
		return m, tea.Quit
	case *exec.Cmd:
		m.hidden = true
		m.exitCmd = msg
		return m, tea.Quit
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

	return lipgloss.NewStyle().Padding(m.options.Padding).Render(pageView)
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
	return m.width - 2*m.options.Padding
}

func (m *Model) pageHeight() int {
	if m.options.MaxHeight > 0 {
		return utils.Min(m.options.MaxHeight, m.height) - 2*m.options.Padding
	}
	return m.height - 2*m.options.Padding
}

type popMsg struct{}

func PopCmd() tea.Msg {
	return popMsg{}
}

type pushMsg struct {
	container Page
}

func NewPushCmd(page Page) tea.Cmd {
	return func() tea.Msg {
		return pushMsg{page}
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

type exit struct {
}

func Exit() tea.Msg { return exit{} }

func Draw(model *Model, fullscreen bool) (err error) {
	// Background detection before we start the program
	lipgloss.SetHasDarkBackground(lipgloss.HasDarkBackground())

	var p *tea.Program
	if fullscreen {
		p = tea.NewProgram(model, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(model)
	}

	m, err := p.Run()
	if err != nil {
		return err
	}

	model, ok := m.(*Model)
	if !ok {
		return fmt.Errorf("could not convert model to *Model")
	}

	if model.exitCmd != nil {
		model.exitCmd.Stdin = os.Stdin
		model.exitCmd.Stdout = os.Stdout
		model.exitCmd.Stderr = os.Stderr

		return model.exitCmd.Run()
	}

	return nil
}
