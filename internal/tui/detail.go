package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Detail struct {
	actionsFocused bool
	width, height  int

	statusBar StatusBar
	Style     lipgloss.Style
	viewport  viewport.Model

	markdown string
}

func NewDetail(markdown string, actions ...types.Action) *Detail {
	viewport := viewport.New(0, 0)
	viewport.Style = lipgloss.NewStyle().Padding(0, 1)

	var statusBar StatusBar
	if len(actions) == 0 {
		statusBar = NewStatusBar()
	} else {
		statusBar = NewStatusBar(actions...)
	}

	items := make([]FilterItem, 0)
	for _, action := range actions {
		items = append(items, ListItem{
			Title:    action.Title,
			Subtitle: action.Key,
			Actions:  []types.Action{action},
		})
	}

	filter := NewFilter(items...)
	filter.DrawLines = true

	d := Detail{
		viewport:  viewport,
		statusBar: statusBar,
		markdown:  markdown,
	}

	_ = d.RefreshContent()
	return &d
}

func (d *Detail) Init() tea.Cmd {
	return d.statusBar.Init()
}

func (d *Detail) Focus() tea.Cmd {
	return nil
}

func (d *Detail) Blur() tea.Cmd {
	return nil
}

type DetailMsg string

func (d *Detail) SetIsLoading(isLoading bool) tea.Cmd {
	return d.statusBar.SetIsLoading(isLoading)
}

func (c *Detail) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if c.actionsFocused {
				break
			}

			return c, func() tea.Msg {
				return ExitMsg{}
			}
		case "esc":
			if c.statusBar.expanded {
				break
			}
			return c, func() tea.Msg {
				return PopPageMsg{}
			}
		}
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd

	c.viewport, cmd = c.viewport.Update(msg)
	cmds = append(cmds, cmd)

	c.statusBar, cmd = c.statusBar.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *Detail) RefreshContent() error {
	render, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(c.width-2),
	)
	if err != nil {
		return err
	}

	content, err := render.Render(c.markdown)
	if err != nil {
		return err
	}

	c.viewport.SetContent(content)
	return nil
}

func (c *Detail) SetSize(width, height int) {
	c.width, c.height = width, height

	c.viewport.Height = height - lipgloss.Height(c.statusBar.View())
	c.viewport.Width = width

	c.statusBar.Width = width
	_ = c.RefreshContent()
}

func (c *Detail) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.viewport.View(), c.statusBar.View())
}
