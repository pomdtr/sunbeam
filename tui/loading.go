package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Loading struct {
	width, height int
	spinner       spinner.Model

	footer Footer
}

func NewLoading() *Loading {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	footer := NewFooter()

	return &Loading{spinner: s, footer: footer}
}

func (c *Loading) SetSize(width, height int) {
	c.footer.Width = width
	c.width = width

	c.height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footer.View())
}

func (c *Loading) Init() tea.Cmd {
	return c.spinner.Tick
}

func (c *Loading) headerView() string {
	loadingText := DefaultStyles.Secondary.Render("Loading...")
	headerRow := fmt.Sprintf("  %s %s", c.spinner.View(), loadingText)
	separator := strings.Repeat("─", c.width)
	return lipgloss.JoinVertical(lipgloss.Left, headerRow, separator)
}

func (c *Loading) Update(msg tea.Msg) (*Loading, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return c, PopCmd
		}
	default:
		var cmd tea.Cmd
		c.spinner, cmd = c.spinner.Update(msg)
		return c, cmd
	}

	return c, nil
}

func (c *Loading) View() string {

	rows := make([]string, c.height)

	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), strings.Join(rows, "\n"), c.footer.View())
}
