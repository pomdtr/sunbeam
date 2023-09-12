package internal

import (
	"encoding/json"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg"
)

type List struct {
	header     Header
	footer     Footer
	actionList ActionList
	filter     Filter

	extension Extension
	command   pkg.Command
	params    pkg.CommandParams

	page *pkg.List
}

func NewList(extension Extension, command pkg.Command, params pkg.CommandParams) *List {
	list := List{
		header: NewHeader(),
		footer: NewFooter(command.Title),

		extension: extension,
		command:   command,
		params:    params,
	}

	list.actionList = NewActionList()

	filter := NewFilter()
	filter.DrawLines = true

	list.filter = filter

	return &list
}

func (c *List) Init() tea.Cmd {
	return tea.Batch(FocusCmd, c.header.SetIsLoading(true), c.Reload)
}

func (c *List) Reload() tea.Msg {
	output, err := c.extension.Run(c.command.Name, pkg.CommandInput{
		Params: c.params,
	})
	if err != nil {
		return err
	}

	if err := pkg.ValidatePage(output); err != nil {
		return err
	}

	var list pkg.List
	if err := json.Unmarshal(output, &list); err != nil {
		return err
	}

	return list
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
	if l.page == nil {
		return nil
	}

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

	if l.page.EmptyView == nil {
		return nil
	}

	actions := l.page.EmptyView.Actions
	l.actionList.SetTitle("Empty Actions")
	l.actionList.SetActions(l.page.EmptyView.Actions...)

	if len(l.page.EmptyView.Actions) > 0 {
		l.footer.SetBindings(
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", actions[0].Title)),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Actions")),
		)
	}

	return nil
}

func (c *List) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case pkg.List:
		c.page = &msg
		c.SetIsLoading(false)

		items := make([]ListItem, len(msg.Items))
		for i, item := range c.page.Items {
			items[i] = ListItem(item)
		}

		c.SetItems(items, "")
		if c.page.Title != "" {
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

			return c, nil
		case "enter":
			if c.actionList.Focused() {
				break
			}

			var actions []pkg.Action
			if c.filter.Selection() != nil {
				item := c.filter.Selection().(ListItem)
				actions = item.Actions
			} else {
				actions = c.page.EmptyView.Actions
			}

			if len(actions) == 0 {
				break
			}

			return c, runAction(actions[0], c.extension)
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
