package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type List struct {
	statusBar     StatusBar
	filter        Filter
	height        int
	width         int
	OnQueryChange func(string) tea.Cmd
}

type QueryChangeMsg string

func NewList(items ...types.ListItem) *List {
	filter := NewFilter()
	filter.Reversed = true
	filter.DrawLines = true

	statusBar := NewStatusBar()
	statusBar.ShowInput()

	list := &List{
		statusBar: statusBar,
		filter:    filter,
	}

	list.SetItems(items...)
	return list
}

func (c *List) Init() tea.Cmd {
	return tea.Batch()
}

func (c *List) Focus() tea.Cmd {
	return c.statusBar.Focus()
}

func (c *List) Blur() tea.Cmd {
	return nil
}

func (c *List) SetQuery(query string) {
	c.statusBar.input.SetValue(query)
}

func (c *List) SetSize(width, height int) {
	c.width, c.height = width, height

	availableHeight := max(0, height-lipgloss.Height(c.statusBar.View()))

	c.statusBar.Width = width
	c.filter.SetSize(width, availableHeight)
}

func (c List) Selection() (types.ListItem, bool) {
	selection := c.filter.Selection()
	if selection == nil {
		return types.ListItem{}, false
	}

	item := selection.(ListItem)
	return types.ListItem(item), true
}

func (c *List) SetItems(items ...types.ListItem) {
	filterItems := make([]FilterItem, len(items))
	for i, item := range items {
		filterItems[i] = ListItem(item)
	}

	c.filter.SetItems(filterItems...)
	c.filter.FilterItems(c.Query())

	selection := c.filter.Selection()
	if selection == nil {
		c.statusBar.SetActions()
	} else {
		c.statusBar.SetActions(selection.(ListItem).Actions...)
	}
}

func (c *List) SetIsLoading(isLoading bool) tea.Cmd {
	return c.statusBar.SetIsLoading(isLoading)
}

func (c *List) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case QueryChangeMsg:
		if c.OnQueryChange == nil {
			return c, nil
		}

		if string(msg) != c.statusBar.Value() {
			return c, nil
		}

		return c, c.OnQueryChange(string(msg))
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	statusBar, cmd := c.statusBar.Update(msg)
	cmds = append(cmds, cmd)
	if statusBar.Value() != c.statusBar.Value() {
		if c.OnQueryChange != nil {
			query := statusBar.Value()
			cmds = append(cmds, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				if query == c.statusBar.Value() {
					return QueryChangeMsg(query)
				}

				return nil
			}))
		} else {
			c.filter.FilterItems(statusBar.Value())
		}
		if c.filter.Selection() != nil {
			statusBar.SetActions(c.filter.Selection().(ListItem).Actions...)
		}
	}

	filter, cmd := c.filter.Update(msg)
	cmds = append(cmds, cmd)
	oldSelection := c.filter.Selection()
	newSelection := filter.Selection()
	if newSelection == nil {
		statusBar.SetActions()
	} else if oldSelection == nil || oldSelection.ID() != newSelection.ID() {
		statusBar.SetActions(newSelection.(ListItem).Actions...)
	}

	c.filter = filter
	c.statusBar = statusBar

	return c, tea.Batch(cmds...)
}

func (c List) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.filter.View(), c.statusBar.View())
}

func (c List) Query() string {
	return c.statusBar.input.Value()
}
