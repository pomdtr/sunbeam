package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/sunbeam"
)

type Container interface {
	SetSize(width, height int)
	Init() tea.Cmd
	Update(tea.Msg) (Container, tea.Cmd)
	View() string
}

type RootModel struct {
	width, height int
	pages         []Container
}

func NewRootModel(rootPage Container) *RootModel {
	return &RootModel{pages: []Container{rootPage}}
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
	case PushMsg:
		m.Push(msg.Page)
		return m, msg.Page.Init()
	case popMsg:
		m.Pop()
		if len(m.pages) == 0 {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	currentPageIdx := len(m.pages) - 1
	m.pages[currentPageIdx], cmd = m.pages[currentPageIdx].Update(msg)

	return m, cmd
}

func (m *RootModel) View() string {
	if len(m.pages) == 0 {
		return ""
	}
	return m.pages[len(m.pages)-1].View()
}

func (m *RootModel) Push(page Container) {
	page.SetSize(m.width, m.height)
	m.pages = append(m.pages, page)
}

func (m *RootModel) Pop() {
	if len(m.pages) > 0 {
		m.pages = m.pages[:len(m.pages)-1]
	}
}

func (m *RootModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	for _, page := range m.pages {
		page.SetSize(width, height)
	}
}

type PushMsg struct {
	Page Container
}

func NewPushCmd(page Container) tea.Cmd {
	return func() tea.Msg {
		return PushMsg{Page: page}
	}
}

type ReplaceMsg struct {
	Page Container
}

func NewReplaceCmd(page Container) tea.Cmd {
	return func() tea.Msg {
		return ReplaceMsg{Page: page}
	}
}

type popMsg struct{}

func PopCmd() tea.Msg {
	return popMsg{}
}

func Draw() error {
	rootItems := make([]sunbeam.ListItem, len(sunbeam.Commands))

	for i, command := range sunbeam.Commands {
		var primaryAction sunbeam.ScriptAction
		if command.Mode == "action" {
			primaryAction = command.Action
			primaryAction.Root = command.Root.String()
			log.Println(primaryAction)
		} else {
			primaryAction = sunbeam.ScriptAction{
				Type:   "push",
				Target: command.Id,
				Root:   command.Root.String(),
			}
		}

		rootItems[i] = sunbeam.ListItem{
			Title:    command.Title,
			Subtitle: command.Subtitle,
			Actions: []sunbeam.ScriptAction{
				primaryAction,
			},
		}
	}

	rootContainer := NewListContainer("Commands", rootItems, RunAction)
	m := NewRootModel(rootContainer)
	p := tea.NewProgram(m, tea.WithAltScreen())
	err := p.Start()

	if err != nil {
		return err
	}
	return nil
}
