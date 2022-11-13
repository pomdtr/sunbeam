package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type Detail struct {
	header     Header
	viewport   viewport.Model
	actionList *ActionList
	footer     Footer
}

func NewDetail(title string, actions ...Action) *Detail {
	viewport := viewport.New(0, 0)
	viewport.Style = styles.Text.Copy().Padding(1, 2)

	footer := NewFooter(title)

	actionList := NewActionList()
	actionList.SetTitle(title)
	actionList.SetActions(actions...)

	if len(actions) > 0 {
		footer.SetBindings(
			actions[0].Binding(),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Show Actions")),
		)
	}

	header := NewHeader()

	return &Detail{
		viewport:   viewport,
		header:     header,
		actionList: actionList,
		footer:     footer,
	}
}

func (d *Detail) Init() tea.Cmd {
	return d.header.SetIsLoading(true)
}

func (d *Detail) SetContent(content string) {
	content = wordwrap.String(content, d.viewport.Width-4)
	d.viewport.SetContent(content)
}

func (c Detail) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			switch msg.String() {
			case "q", "Q":
				return &c, tea.Quit
			}
		case tea.KeyEscape:
			return &c, PopCmd
		}
	}
	var cmd tea.Cmd

	c.viewport, cmd = c.viewport.Update(msg)

	return &c, cmd
}

func (c *Detail) SetSize(width, height int) {
	c.footer.Width = width
	c.header.Width = width
	c.viewport.Width = width
	c.viewport.Height = height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())
}

func (c *Detail) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
}
