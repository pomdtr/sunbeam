package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/chroma/quick"
	"github.com/pomdtr/sunbeam/types"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/pomdtr/sunbeam/utils"
)

const debounceDuration = 300 * time.Millisecond

// Probably not necessary, need to be refactored
type ListItem struct {
	types.ListItem
}

func ParseScriptItem(scriptItem types.ListItem) ListItem {
	return ListItem{
		ListItem: scriptItem,
	}
}

func (i ListItem) ID() string {
	if i.Id != "" {
		return i.Id
	}
	return i.Title
}

func (i ListItem) FilterValue() string {
	keywords := []string{i.Title, i.Subtitle}
	return strings.Join(keywords, " ")
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
	header       Header
	footer       Footer
	actionList   ActionList
	emptyActions []types.Action
	DetailFunc   func(types.ListItem) string

	ShowPreview bool

	filter         Filter
	viewport       viewport.Model
	previewContent string
}

func NewList(page *types.Page) *List {
	viewport := viewport.New(0, 0)
	viewport.Style = lipgloss.NewStyle().Padding(0, 1)

	list := List{
		actionList:  NewActionList(),
		header:      NewHeader(),
		ShowPreview: page.ShowPreview,
		viewport:    viewport,
		footer:      NewFooter(page.Title),
	}

	filter := NewFilter()
	filter.DrawLines = true

	if page.EmptyView != nil {
		filter.EmptyText = page.EmptyView.Text
		list.emptyActions = page.EmptyView.Actions
	}

	list.filter = filter

	list.DetailFunc = func(item types.ListItem) string {
		if item.Preview == nil {
			return ""
		}

		if item.Preview.Text != "" {
			if item.Preview.HighLight == "" {
				return item.Preview.Text
			}
			builder := strings.Builder{}
			if err := quick.Highlight(&builder, item.Preview.Text, item.Preview.HighLight, "terminal16", "github"); err != nil {
				return item.Preview.Text
			}
			return builder.String()
		}

		output, err := item.Preview.Command.Output()
		if err != nil {
			return err.Error()
		}

		content := string(output)

		if item.Preview.HighLight == "" {
			return content
		}

		builder := strings.Builder{}
		if err := quick.Highlight(&builder, content, item.Preview.HighLight, "terminal16", "github"); err != nil {
			return content
		}
		return builder.String()
	}

	return &list
}

func (c *List) Init() tea.Cmd {
	return c.header.Focus()
}

func (c *List) RefreshDetail() {
	c.viewport.SetYOffset(0)
	detailWidth := c.viewport.Width - 2 // take padding into account
	detailContent := wrap.String(wordwrap.String(c.previewContent, detailWidth), detailWidth)

	c.viewport.SetContent(detailContent)
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
		c.RefreshDetail()
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

func (c *List) SetItems(items []ListItem, selectedId string) tea.Cmd {
	filterItems := make([]FilterItem, len(items))
	for i, item := range items {
		filterItems[i] = item
	}

	c.filter.SetItems(filterItems)
	c.filter.FilterItems(c.Query())
	if selectedId != "" {
		c.filter.Select(selectedId)
	}

	selection := c.updateSelection(c.filter)

	if selection == nil {
		return nil
	}

	if c.filter.Selection() == nil {
		return nil
	}

	return func() tea.Msg {
		return SelectionChangeMsg{
			SelectionId: c.filter.Selection().ID(),
		}
	}
}

func (c *List) SetIsLoading(isLoading bool) tea.Cmd {
	return c.header.SetIsLoading(isLoading)
}

type ContentMsg string

func (l *List) updateSelection(filter Filter) FilterItem {
	if filter.Selection() == nil {
		l.previewContent = ""
		l.actionList.SetTitle("Empty Actions")
		l.actionList.SetActions(l.emptyActions...)

		if len(l.emptyActions) > 0 {
			l.footer.SetBindings(
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", l.emptyActions[0].Title)),
				key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Show Actions")),
			)
		}
	} else {
		item := filter.Selection().(ListItem)
		l.actionList.SetTitle(item.Title)
		l.actionList.SetActions(item.Actions...)
		if len(item.Actions) > 0 {
			l.footer.SetBindings(
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", item.Actions[0].Title)),
				key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Show Actions")),
			)
		}
	}

	return l.filter.Selection()
}

func (c *List) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if c.actionList.Focused() {
				break
			} else if c.header.input.Value() != "" {
				c.header.input.SetValue("")
				c.filter.FilterItems("")
				c.updateSelection(c.filter)
				selection := c.filter.Selection()
				if selection == nil {
					return c, nil
				}
				return c, tea.Sequence(func() tea.Msg {
					return SelectionChangeMsg{
						SelectionId: selection.ID(),
					}
				})
			} else {
				return c, func() tea.Msg {
					return PopPageMsg{}
				}
			}
		case "tab":
			if c.actionList.Focused() && len(c.actionList.actions) > 0 {
				break
			}

			if c.filter.Selection() == nil {
				break
			}

			item := c.filter.Selection().(ListItem)
			if len(item.Actions) == 1 {
				break
			}

			return c, c.actionList.Focus()
		case "shift+down":
			c.viewport.LineDown(1)
			return c, nil
		case "shift+up":
			c.viewport.LineUp(1)
			return c, nil
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
		if c.DetailFunc == nil {
			return c, nil
		}

		cmd := c.SetIsLoading(true)

		return c, tea.Sequence(cmd, func() tea.Msg {
			return ContentMsg(c.DetailFunc(item.ListItem))
		})

	case ContentMsg:
		c.SetIsLoading(false)
		c.previewContent = string(msg)
		c.RefreshDetail()
		return c, nil
	}

	header, cmd := c.header.Update(msg)
	cmds = append(cmds, cmd)

	filter, cmd := c.filter.Update(msg)
	cmds = append(cmds, cmd)

	c.actionList, cmd = c.actionList.Update(msg)
	cmds = append(cmds, cmd)

	if c.actionList.Focused() {
		return c, tea.Batch(cmds...)
	}

	if header.Value() != c.header.Value() {
		filter.FilterItems(header.Value())
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

func (c *List) SetQuery(query string) {
	c.header.input.SetValue(query)
}

func (c List) Query() string {
	return c.header.input.Value()
}
