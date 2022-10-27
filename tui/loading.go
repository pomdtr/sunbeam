package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/utils"
)

type Loading struct {
	width, height int
	spinner       spinner.Model
}

func NewLoading() *Loading {
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &Loading{spinner: s}
}

func (c *Loading) headerView() string {
	line := strings.Repeat("─", c.width)
	return fmt.Sprintf("\n%s", line)
}

func (c *Loading) SetSize(width, height int) {
	c.width = width
	c.height = height
}

func (c *Loading) footerView() string {
	return SunbeamFooter(c.width, "")
}

func (c *Loading) Init() tea.Cmd {
	return c.spinner.Tick
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
	log.Println(c.width, c.height)
	var loadingIndicator string
	spinner := lipgloss.NewStyle().Padding(0, 2).Render(c.spinner.View())
	label := lipgloss.NewStyle().Render("Loading...")
	loadingIndicator = lipgloss.JoinHorizontal(lipgloss.Center, spinner, label)

	newLines := strings.Repeat("\n", utils.Max(0, c.height-lipgloss.Height(loadingIndicator)-lipgloss.Height(c.footerView())-lipgloss.Height(c.headerView())-1))

	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), loadingIndicator, newLines, c.footerView())
}
