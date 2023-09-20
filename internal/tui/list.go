package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type List struct {
	header     Header
	filter     Filter
	actionList ActionList
	footer     Footer
}

func NewList(title string, items ...types.ListItem) *List {
	filter := NewFilter()
	filter.DrawLines = true

	list := &List{
		header:     NewHeader(),
		actionList: NewActionList(),
		footer:     NewFooter(title),
		filter:     filter,
	}

	list.SetItems(items...)
	return list
}

func (c *List) Init() tea.Cmd {
	return tea.Batch(FocusCmd)
}

func (c *List) SetSize(width, height int) {
	availableHeight := max(0, height-lipgloss.Height(c.header.View())-lipgloss.Height(c.footer.View()))

	c.footer.Width = width
	c.header.Width = width
	c.actionList.SetSize(width, height)
	c.filter.SetSize(width, availableHeight)
}

func (c List) Selection() *ListItem {
	if c.filter.Selection() == nil {
		return nil
	}
	item := c.filter.Selection().(ListItem)

	return &item
}

func (c *List) SetItems(items ...types.ListItem) {
	filterItems := make([]FilterItem, len(items))
	for i, item := range items {
		filterItems[i] = ListItem(item)
	}

	c.filter.SetItems(filterItems...)
	c.filter.FilterItems(c.Query())
	c.updateSelection(c.filter)
}

func (c *List) SetIsLoading(isLoading bool) tea.Cmd {
	return c.header.SetIsLoading(isLoading)
}

func (l *List) updateSelection(filter Filter) FilterItem {
	if filter.Selection() != nil {
		item := filter.Selection().(ListItem)
		l.actionList.SetTitle(item.Title)
		l.actionList.SetActions(item.Actions...)
		if len(item.Actions) > 0 {
			l.footer.SetBindings(
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", item.Actions[0].Title)),
				key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Actions")),
			)
		}

		return filter.Selection()
	}

	return nil
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
				return c, nil
			} else {
				return c, func() tea.Msg {
					return PopPageMsg{}
				}
			}
		case "tab":
			if c.actionList.Focused() {
				break
			}

			if c.filter.Selection() == nil {
				return c, nil
			}

			item := c.filter.Selection().(ListItem)
			if len(item.Actions) == 0 {
				return c, nil
			}

			return c, c.actionList.Focus()
		case "enter":
			if c.actionList.Focused() {
				break
			}

			return c, func() tea.Msg {
				selection := c.filter.Selection()
				if selection == nil {
					return nil
				}

				item := c.filter.Selection().(ListItem)
				if len(item.Actions) == 0 {
					return nil
				}

				return item.Actions[0]
			}
		}
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
		c.updateSelection(filter)
	}

	c.header = header
	c.filter = filter
	c.updateSelection(c.filter)

	return c, tea.Batch(cmds...)
}

func (c List) View() string {
	if c.actionList.Focused() {
		return c.actionList.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.filter.View(), c.footer.View())
}

func (c *List) SetQuery(query string) {
	c.header.input.SetValue(query)
}

func (c List) Query() string {
	return c.header.input.Value()
}
