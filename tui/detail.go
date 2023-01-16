package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sunbeamlauncher/sunbeam/app"
	"github.com/sunbeamlauncher/sunbeam/utils"
)

type Detail struct {
	header       Header
	Style        lipgloss.Style
	content      string
	metadatas    []app.ScriptMetadata
	mainViewport viewport.Model
	sideViewport viewport.Model
	actionList   ActionList
	footer       Footer
}

func NewDetail(title string) *Detail {
	mainViewport := viewport.New(0, 0)
	sideViewport := viewport.New(0, 0)

	footer := NewFooter(title)

	actionList := NewActionList()
	actionList.SetTitle(title)

	header := NewHeader()

	d := Detail{
		mainViewport: mainViewport,
		sideViewport: sideViewport,
		header:       header,
		actionList:   actionList,
		footer:       footer,
	}

	return &d
}

func (c *Detail) SetActions(actions ...Action) {
	c.actionList.SetActions(actions...)

	if len(actions) == 0 {
		c.footer.SetBindings()
	} else {
		c.footer.SetBindings(
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", actions[0].Title)),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Show Actions")),
		)
	}
}

func (d *Detail) Init() tea.Cmd {
	return nil
}

func (d *Detail) updateContent() {
	mainContent := wordwrap.String(d.content, utils.Max(0, d.mainViewport.Width-2))
	mainContent = lipgloss.NewStyle().Padding(0, 1).Width(d.mainViewport.Width).Render(mainContent)
	d.mainViewport.SetContent(mainContent)

	items := make([]string, 0)
	maxWidth := utils.Max(d.sideViewport.Width-2, 0)
	for _, metadata := range d.metadatas {
		items = append(items, fmt.Sprintf(
			"%s\n%s",
			styles.Faint.MaxWidth(maxWidth).Render(metadata.Title),
			lipgloss.NewStyle().MaxWidth(maxWidth).Render(metadata.Value),
		))
	}

	sideContent := strings.Join(items, "\n\n")
	sideContent = lipgloss.NewStyle().Padding(0, 1).Width(maxWidth).Render(sideContent)
	d.sideViewport.SetContent(sideContent)
}

type DetailMsg string

func (d *Detail) SetContent(content string) {
	d.content = content
	d.updateContent()
}

func (d *Detail) SetDetail(detail app.Detail) tea.Cmd {
	actions := make([]Action, len(detail.Actions))
	for i, action := range detail.Actions {
		actions[i] = NewAction(action)
	}

	d.SetActions(actions...)
	d.content = detail.Preview
	d.metadatas = detail.Metadatas
	d.updateContent()

	return nil
}

func (d *Detail) SetIsLoading(isLoading bool) tea.Cmd {
	return d.header.SetIsLoading(isLoading)
}

func (c Detail) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			switch msg.String() {
			case "q", "Q":
				return &c, tea.Quit
			}
		case tea.KeyEscape:
			if c.actionList.Focused() {
				break
			}
			return &c, PopCmd
		case tea.KeyShiftDown:
			c.sideViewport.LineDown(1)
			return &c, nil
		case tea.KeyShiftUp:
			c.sideViewport.LineUp(1)
			return &c, nil
		}
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd

	c.mainViewport, cmd = c.mainViewport.Update(msg)
	cmds = append(cmds, cmd)

	c.actionList, cmd = c.actionList.Update(msg)
	cmds = append(cmds, cmd)

	c.header, cmd = c.header.Update(msg)
	cmds = append(cmds, cmd)

	return &c, tea.Batch(cmds...)
}

func (c Detail) SideBarVisible() bool {
	return len(c.metadatas) > 0
}

func (c *Detail) SetSize(width, height int) {
	c.footer.Width = width
	c.header.Width = width
	c.actionList.SetSize(width, height)

	viewportHeight := height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())
	if c.SideBarVisible() {
		metadataWidth := width / 3

		c.mainViewport.Width = width - metadataWidth - 1
		c.mainViewport.Height = viewportHeight

		c.sideViewport.Width = metadataWidth
		c.sideViewport.Height = viewportHeight
	} else {
		c.mainViewport.Width = width
		c.mainViewport.Height = viewportHeight
	}

	c.updateContent()
}

func (c *Detail) View() string {
	if c.actionList.Focused() {
		return c.actionList.View()
	}

	if c.SideBarVisible() {
		var separatorChars = make([]string, c.mainViewport.Height)
		for i := 0; i < c.mainViewport.Height; i++ {
			separatorChars[i] = "│"
		}
		separator := strings.Join(separatorChars, "\n")

		view := lipgloss.JoinHorizontal(lipgloss.Top,
			c.mainViewport.View(),
			separator,
			c.sideViewport.View(),
		)

		return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), view, c.footer.View())
	}
	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.mainViewport.View(), c.footer.View())
}
