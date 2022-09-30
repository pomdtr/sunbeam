package pages

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

type NewSelectActionCmd func(commands.ScriptAction) tea.Cmd

type DetailContainer struct {
	response     commands.DetailResponse
	selectAction NewSelectActionCmd
	viewport     *viewport.Model
}

func NewDetailContainer(response *commands.DetailResponse, selectAction NewSelectActionCmd) DetailContainer {
	viewport := viewport.New(0, 0)
	var content string
	if lipgloss.HasDarkBackground() {
		content, _ = glamour.Render(response.Text, "dark")
	} else {
		content, _ = glamour.Render(response.Text, "light")
	}
	viewport.SetContent(content)

	return DetailContainer{
		response:     *response,
		selectAction: selectAction,
		viewport:     &viewport,
	}
}

func (c DetailContainer) SetSize(width, height int) {
	c.viewport.Width = width
	c.viewport.Height = height - lipgloss.Height(c.footerView())
}

func (c DetailContainer) Init() tea.Cmd {
	return nil
}

func (m DetailContainer) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", utils.Max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (c DetailContainer) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			for _, action := range c.response.Actions {
				if action.Keybind == string(msg.Runes) {
					return c, c.selectAction(action)
				}
			}
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
	return fmt.Sprintf("%s\n%s", c.viewport.View(), c.footerView())
}
