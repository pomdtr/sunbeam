package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Detail struct {
	actionsFocused bool

	isLoading bool
	spinner   spinner.Model
	viewport  viewport.Model
	statusBar StatusBar

	text          string
	width, height int

	Style     lipgloss.Style
	Highlight types.Highlight
}

func NewDetail(text string, actions ...types.Action) *Detail {
	viewport := viewport.New(0, 0)
	viewport.Style = lipgloss.NewStyle()

	statusBar := NewStatusBar(actions...)
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
		spinner:   spinner.New(),
		viewport:  viewport,
		statusBar: statusBar,
		Highlight: types.HighlightAnsi,
		text:      text,
	}

	_ = d.RefreshContent()
	return &d
}

func (d *Detail) Init() tea.Cmd {
	return nil
}

func (d *Detail) Focus() tea.Cmd {
	return nil
}

func (d *Detail) Blur() tea.Cmd {
	return nil
}

type DetailMsg string

func (d *Detail) SetIsLoading(isLoading bool) tea.Cmd {
	d.isLoading = isLoading
	if isLoading {
		return d.spinner.Tick
	}

	return nil
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
				c.statusBar.expanded = false
				c.statusBar.cursor = 0
				return c, nil
			}
			return c, func() tea.Msg {
				return PopPageMsg{}
			}
		}
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if c.isLoading {
		c.spinner, cmd = c.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	c.viewport, cmd = c.viewport.Update(msg)
	cmds = append(cmds, cmd)

	c.statusBar, cmd = c.statusBar.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *Detail) RefreshContent() error {
	var content string
	if c.Highlight == types.HighlightMarkdown {
		render, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(c.width),
		)
		if err != nil {
			return err
		}

		content, err = render.Render(c.text)
		if err != nil {
			return err
		}
	} else {
		content = wordwrap.String(c.text, c.width)
	}

	c.viewport.SetContent(content)
	return nil
}

func (c *Detail) SetSize(width, height int) {
	c.width, c.height = width, height

	c.viewport.Height = height - 4
	c.viewport.Width = width

	c.statusBar.Width = width
	_ = c.RefreshContent()
}

func (c *Detail) View() string {
	var headerRow string
	if c.isLoading {
		headerRow = " " + c.spinner.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, headerRow, separator(c.width), c.viewport.View(), c.statusBar.View())
}
