package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var markdownRenderer *glamour.TermRenderer

func init() {
	var err error
	markdownRenderer, err = glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		log.Fatalf("failed to create markdown renderer: %v", err)
	}
}

type Detail struct {
	format  string
	spinner spinner.Model

	header Header
	viewport.Model
	actionList *ActionList
	footer     *Footer
}

func NewDetail(title string, actions ...Action) *Detail {
	spinner := spinner.New()
	viewport := viewport.New(0, 0)

	footer := NewFooter(title)
	footer.SetBindings(
		actions[0].Binding(),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Show Actions")),
	)

	actionList := NewActionList()
	actionList.SetTitle(title)
	actionList.SetActions(actions...)

	header := NewHeader()

	return &Detail{Model: viewport, header: header, spinner: spinner, actionList: actionList, format: "raw", footer: footer}
}

func (d *Detail) Init() tea.Cmd {
	return d.spinner.Tick
}

func (d *Detail) SetContent(content string) error {
	switch d.format {
	case "markdown":
		rendered, err := markdownRenderer.Render(content)
		if err != nil {
			return err
		}
		d.Model.SetContent(rendered)
	default:
		d.Model.SetContent(content)
	}
	return nil
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
	var cmds []tea.Cmd

	c.spinner, cmd = c.spinner.Update(msg)
	cmds = append(cmds, cmd)

	c.Model, cmd = c.Model.Update(msg)
	cmds = append(cmds, cmd)

	return &c, tea.Batch(cmds...)
}

func (c *Detail) SetSize(width, height int) {
	c.footer.Width = width
	c.header.Width = width
	c.Model.Width = width
	c.Model.Height = height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())
}

func (c *Detail) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.Model.View(), c.footer.View())
}
