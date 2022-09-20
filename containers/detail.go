package containers

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	commands "github.com/pomdtr/sunbeam/commands"
	"github.com/pomdtr/sunbeam/utils"
)

var infoStyle = func() lipgloss.Style {
	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	return titleStyle.Copy().BorderStyle(b)
}()

type DetailContainer struct {
	command  commands.Command
	viewport *viewport.Model
}

func NewDetailContainer(command commands.Command, markdown string) DetailContainer {
	viewport := viewport.New(0, 0)
	var content string
	if lipgloss.HasDarkBackground() {
		content, _ = glamour.Render(markdown, "dark")
	} else {
		content, _ = glamour.Render(markdown, "light")
	}
	viewport.SetContent(content)

	return DetailContainer{
		command:  command,
		viewport: &viewport,
	}
}

func (c DetailContainer) SetSize(width, height int) {
	c.viewport.Width = width
	c.viewport.Height = height - 2
}

func (c DetailContainer) Init() tea.Cmd {
	return nil
}

func (m DetailContainer) headerView() string {
	title := titleStyle.Render("Mr. Pager")
	line := strings.Repeat("─", utils.Max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m DetailContainer) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", utils.Max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (c DetailContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return c, PopCmd
		}
	}
	var cmd tea.Cmd
	model, cmd := c.viewport.Update(msg)
	c.viewport = &model
	return c, cmd
}

func (c DetailContainer) View() string {
	return fmt.Sprintf("%s\n%s\n%s", c.headerView(), c.viewport.View(), c.footerView())
}
