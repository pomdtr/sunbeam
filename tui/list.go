package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/utils"
)

const debounceDuration = 300 * time.Millisecond

type ListItem struct {
	Id          string
	Title       string
	Subtitle    string
	PreviewFunc func() string
	Accessories []string
	Actions     []Action
}

func ParseScriptItem(scriptItem schemas.ListItem, dir string) ListItem {
	actions := make([]Action, len(scriptItem.Actions))
	for i, scriptAction := range scriptItem.Actions {
		actions[i] = NewAction(scriptAction, dir)
	}

	return ListItem{
		Id:          scriptItem.Id,
		Title:       scriptItem.Title,
		Subtitle:    scriptItem.Subtitle,
		Accessories: scriptItem.Accessories,
		Actions:     actions,
	}

}

func (i ListItem) ID() string {
	return i.Id
}

func (i ListItem) FilterValue() string {
	if i.Subtitle == "" {
		return i.Title
	}
	return fmt.Sprintf("%s %s", i.Title, i.Subtitle)
}

func (i ListItem) Render(width int, selected bool) string {
	if width == 0 {
		return ""
	}

	var title string
	titleStyle := lipgloss.NewStyle().Bold(true)
	if selected {
		title = fmt.Sprintf("> %s", i.Title)
		titleStyle = titleStyle.Foreground(lipgloss.Color("13"))
	} else {
		title = fmt.Sprintf("  %s", i.Title)
	}

	subtitle := fmt.Sprintf(" %s", i.Subtitle)
	var blanks string
	accessories := fmt.Sprintf(" %s", strings.Join(i.Accessories, " · "))

	// If the width is too small, we need to truncate the subtitle, accessories, or title (in that order)
	if width >= lipgloss.Width(title+subtitle+accessories) {
		availableWidth := width - lipgloss.Width(title+subtitle+accessories)
		blanks = strings.Repeat(" ", availableWidth)
	} else if width >= lipgloss.Width(title+accessories) {
		subtitle = subtitle[:width-lipgloss.Width(title+accessories)]
	} else if width >= lipgloss.Width(title) {
		subtitle = ""
		accessories = accessories[:width-lipgloss.Width(title)]
	} else {
		accessories = ""
		title = title[:utils.Min(len(title), width)]
	}

	title = titleStyle.Render(title)
	subtitle = lipgloss.NewStyle().Faint(true).Render(subtitle)
	accessories = lipgloss.NewStyle().Faint(true).Render(accessories)

	return lipgloss.JoinHorizontal(lipgloss.Top, title, subtitle, blanks, accessories)
}

type List struct {
	header     Header
	footer     Footer
	actionList ActionList
	actions    []Action

	GenerateItems bool
	ShowPreview   bool

	filter         Filter
	viewport       viewport.Model
	previewContent string
}

func (c *List) SetTitle(title string) {
	c.footer.title = title
}

func NewList(title string, actions []Action) *List {
	header := NewHeader()

	viewport := viewport.New(0, 0)
	viewport.Style = lipgloss.NewStyle().Padding(0, 1)
	filter := NewFilter()
	filter.DrawLines = true
	footer := NewFooter(title)

	return &List{
		actions:    actions,
		actionList: NewActionList(),
		header:     header,
		filter:     filter,
		viewport:   viewport,
		footer:     footer,
	}
}

func (c *List) Init() tea.Cmd {
	return c.header.Focus()
}

func (c *List) RefreshPreview() {
	c.viewport.SetYOffset(0)
	previewWidth := c.viewport.Width - 2 // take padding into account
	previewContent := wrap.String(wordwrap.String(c.previewContent, previewWidth), previewWidth)

	c.viewport.SetContent(previewContent)
}

func (c *List) SetSize(width, height int) {
	availableHeight := utils.Max(0, height-lipgloss.Height(c.header.View())-lipgloss.Height(c.footer.View()))
	c.footer.Width = width
	c.header.Width = width
	c.actionList.SetSize(width, height)
	if c.ShowPreview {
		listWidth := width/3 - 1 // take separator into account
		c.filter.SetSize(listWidth, availableHeight)
		c.viewport.Width = width - listWidth
		c.viewport.Height = availableHeight
		c.RefreshPreview()
	} else {
		c.filter.SetSize(width, availableHeight)
	}
}

func (c List) Selection() *ListItem {
	if c.filter.Selection() == nil {
		return nil
	}
	item := c.filter.Selection().(ListItem)

	return &item
}

func (c *List) SetItems(items []ListItem) tea.Cmd {
	filterItems := make([]FilterItem, len(items))
	for i, item := range items {
		filterItems[i] = item
	}

	c.filter.SetItems(filterItems)
	if !c.GenerateItems {
		c.filter.FilterItems(c.Query())
	}

	selection := c.updateSelection(c.filter)

	if selection == nil {
		return nil
	}

	return func() tea.Msg {
		return SelectionChangeMsg{c.filter.Selection().ID()}
	}
}

func (c *List) SetIsLoading(isLoading bool) tea.Cmd {
	return c.header.SetIsLoading(isLoading)
}

type PreviewContentMsg string

func (l *List) updateSelection(filter Filter) FilterItem {
	actions := make([]Action, 0)
	if filter.Selection() == nil {
		l.previewContent = ""
	} else {
		item := filter.Selection().(ListItem)
		l.actionList.SetTitle(item.Title)
		actions = append(actions, item.Actions...)
	}

	for _, action := range l.actions {
		actions = append(actions, action)
	}

	l.actionList.SetActions(actions...)
	if len(actions) > 0 {
		l.footer.SetBindings(
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", actions[0].Title)),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Show Actions")),
		)
	} else {
		l.footer.SetBindings()
	}

	return l.filter.Selection()
}

func (c *List) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	header, cmd := c.header.Update(msg)
	cmds = append(cmds, cmd)

	filter, cmd := c.filter.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			if c.actionList.Focused() {
				break
			} else if c.header.input.Value() != "" {
				header.input.SetValue("")
				filter.FilterItems("")
				c.updateSelection(filter)
			} else {
				return c, func() tea.Msg {
					return PopPageMsg{}
				}
			}
		case tea.KeyShiftDown:
			c.viewport.LineDown(1)
			return c, nil
		case tea.KeyShiftUp:
			c.viewport.LineUp(1)
			return c, nil
		}
	case UpdateQueryMsg:
		if !c.GenerateItems {
			return c, nil
		}
		if c.Query() != msg.Query {
			return c, nil
		}

		return c, func() tea.Msg {
			return ReloadPageMsg{}
		}
	case SelectionChangeMsg:
		if !c.ShowPreview {
			return c, nil
		}

		if c.filter.Selection() == nil {
			return c, nil
		}

		if msg.SelectionId != c.filter.Selection().ID() {
			return c, nil
		}

		item := c.filter.Selection().(ListItem)
		if item.PreviewFunc == nil {
			return c, nil
		}

		cmd := c.SetIsLoading(true)

		return c, tea.Sequence(cmd, func() tea.Msg {
			return PreviewContentMsg(item.PreviewFunc())
		})

	case PreviewContentMsg:
		c.SetIsLoading(false)
		c.previewContent = string(msg)
		c.RefreshPreview()
		return c, nil
	}

	c.actionList, cmd = c.actionList.Update(msg)
	cmds = append(cmds, cmd)

	if c.actionList.Focused() {
		return c, tea.Batch(cmds...)
	}

	if header.Value() != c.header.Value() {
		cmds = append(cmds, tea.Tick(debounceDuration, func(t time.Time) tea.Msg {
			return UpdateQueryMsg{Query: header.Value()}
		}))
		if !c.GenerateItems {
			filter.FilterItems(header.Value())
		}
	}

	if c.filter.Selection() != nil && filter.Selection() != nil && filter.Selection().ID() != c.filter.Selection().ID() {
		cmds = append(cmds, tea.Tick(debounceDuration, func(t time.Time) tea.Msg {
			return SelectionChangeMsg{SelectionId: filter.Selection().ID()}
		}))
		c.updateSelection(filter)
	}

	c.header = header
	c.filter = filter
	c.updateSelection(c.filter)

	return c, tea.Batch(cmds...)
}

type SelectionChangeMsg struct {
	SelectionId string
}

type UpdateQueryMsg struct {
	Query string
}

func (c List) View() string {
	if c.actionList.Focused() {
		return c.actionList.View()
	}

	if c.ShowPreview {
		var separatorChars = make([]string, c.viewport.Height)
		for i := 0; i < c.viewport.Height; i++ {
			separatorChars[i] = "│"
		}
		separator := strings.Join(separatorChars, "\n")
		view := lipgloss.JoinHorizontal(lipgloss.Top, c.filter.View(), separator, c.viewport.View())

		return lipgloss.JoinVertical(lipgloss.Top, c.header.View(), view, c.footer.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.filter.View(), c.footer.View())
}

func (c List) Query() string {
	return c.header.input.Value()
}
