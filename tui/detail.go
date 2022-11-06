package tui

import (
	"fmt"
	"log"
	"strings"

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
	width, height int
	title         string

	viewport.Model
	format    string
	spinner   spinner.Model
	isLoading bool
	actions   []Action
	footer    Footer
}

func NewDetail(title string) *Detail {
	spinner := spinner.New()
	viewport := viewport.New(0, 0)
	footer := NewFooter(title)

	return &Detail{Model: viewport, spinner: spinner, format: "raw", footer: footer}
}

func (d Detail) headerView() string {
	var headerRow string
	if d.isLoading {
		label := DefaultStyles.Secondary.Render("Loading...")
		headerRow = fmt.Sprintf(" %s %s", d.spinner.View(), label)
	} else {
		headerRow = strings.Repeat(" ", d.width)
	}
	separator := strings.Repeat("─", d.Width)
	return lipgloss.JoinVertical(lipgloss.Left, headerRow, separator)
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
		content = lipgloss.NewStyle().Width(d.width).Padding(1, 2).Render(content)
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
	c.width = width
	c.footer.Width = width
	c.Model.Width = width
	c.Model.Height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footer.View())
}

func (c *Detail) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.Model.View(), c.footer.View())
}
