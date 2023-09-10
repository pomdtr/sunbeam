package internal

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg"
	"github.com/pomdtr/sunbeam/utils"
)

type List struct {
	header       Header
	footer       Footer
	actionList   ActionList
	emptyActions []pkg.Action

	generator func() (pkg.List, error)
	runner    func(pkg.Action) tea.Cmd

	viewport      viewport.Model
	filter        Filter
	detailContent string
}

func NewList(title string, generator func() (pkg.List, error), runner func(pkg.Action) tea.Cmd) *List {
	viewport := viewport.New(0, 0)
	list := List{
		header:   NewHeader(),
		footer:   NewFooter(title),
		viewport: viewport,

		generator: generator,
		runner:    runner,
	}

	list.actionList = NewActionList(runner)

	filter := NewFilter()
	filter.DrawLines = true

	list.filter = filter

	return &list
}

func (c *List) Init() tea.Cmd {
	return tea.Batch(c.header.Focus(), c.header.SetIsLoading(true), c.Reload)
}

func (c *List) Reload() tea.Msg {
	list, err := c.generator()
	if err != nil {
		return err
	}

	return list
}

func (c *List) Focus() tea.Cmd {
	return c.header.Focus()
}

func (c *List) SetSize(width, height int) {
	availableHeight := utils.Max(0, height-lipgloss.Height(c.header.View())-lipgloss.Height(c.footer.View()))

	c.footer.Width = width
	c.header.Width = width
	c.actionList.SetSize(width, height)
	c.filter.SetSize(width, availableHeight)
	c.viewport.Width = width
	c.viewport.Height = availableHeight
}

func (c List) Selection() *ListItem {
	if c.filter.Selection() == nil {
		return nil
	}
	item := c.filter.Selection().(ListItem)

	return &item
}

func (c *List) SetItems(items []ListItem, selectedId string) {
	filterItems := make([]FilterItem, len(items))
	for i, item := range items {
		filterItems[i] = item
	}

	c.filter.SetItems(filterItems)
	c.filter.FilterItems(c.Query())
	if selectedId != "" {
		c.filter.Select(selectedId)
	}

	c.updateSelection(c.filter)
}

func (c *List) SetIsLoading(isLoading bool) tea.Cmd {
	return c.header.SetIsLoading(isLoading)
}

func (l *List) updateSelection(filter Filter) FilterItem {
	if filter.Selection() == nil {
		l.detailContent = ""
		l.actionList.SetTitle("Empty Actions")
		l.actionList.SetActions(l.emptyActions...)

		if len(l.emptyActions) > 0 {
			l.footer.SetBindings(
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", l.emptyActions[0].Title)),
				key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Actions")),
			)
		}
	} else {
		item := filter.Selection().(ListItem)
		l.actionList.SetTitle(item.Title)
		l.actionList.SetActions(item.Actions...)
		if len(item.Actions) > 0 {
			l.footer.SetBindings(
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", item.Actions[0].Title)),
				key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Actions")),
			)
		}
	}

	return l.filter.Selection()
}

func (c *List) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case pkg.List:
		c.SetIsLoading(false)
		items := make([]ListItem, len(msg.Items))

		for i, item := range msg.Items {
			items[i] = ListItem(item)
		}

		c.SetItems(items, "")
		if msg.Title != "" {
			c.footer.title = msg.Title
		}
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

			var actions []pkg.Action
			if c.filter.Selection() != nil {
				item := c.filter.Selection().(ListItem)
				actions = item.Actions
			} else {
				actions = c.emptyActions
			}

			if len(actions) == 0 {
				break
			}

			return c, c.runner(actions[0])
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

	if len(c.filter.filtered) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.viewport.View(), c.footer.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.filter.View(), c.footer.View())
}

func (c *List) SetQuery(query string) {
	c.header.input.SetValue(query)
}

func (c List) Query() string {
	return c.header.input.Value()
}
