package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type List struct {
	width, height int

	input textinput.Model

	filter Filter
	footer StatusBar

	Actions       []types.Action
	OnQueryChange func(string) tea.Cmd
	OnSelect      func(string) tea.Cmd
}

type QueryChangeMsg string

func NewList(title string, items ...types.ListItem) *List {
	filter := NewFilter()
	filter.DrawLines = true

	statusBar := NewStatusBar(title)
	input := textinput.New()
	input.Prompt = ""

	list := &List{
		input:  input,
		filter: filter,
		footer: statusBar,
	}

	list.SetItems(items...)
	return list
}

func (l *List) SetActions(actions ...types.Action) {
	l.Actions = actions
	if l.filter.Selection() == nil {
		l.footer.SetActions(actions...)
	}
}

func (l *List) SetEmptyText(text string) {
	l.filter.EmptyText = text
}

func (c *List) Init() tea.Cmd {
	return tea.Batch()
}

func (c *List) Focus() tea.Cmd {
	return c.input.Focus()
}

func (c *List) Blur() tea.Cmd {
	return nil
}

func (c *List) SetQuery(query string) {
	c.input.SetValue(query)
	if c.OnQueryChange == nil {
		c.filter.FilterItems(query)
	}
}

func (c *List) SetSize(width, height int) {
	c.width, c.height = width, height
	availableHeight := max(0, height-4)

	c.footer.Width = width
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
		c.footer.SetActions(c.Actions...)
	} else {
		c.footer.SetActions(selection.(ListItem).Actions...)
	}
}

func (c *List) SetIsLoading(isLoading bool) tea.Cmd {
	return c.footer.SetIsLoading(isLoading)
}

func (c List) Query() string {
	return c.input.Value()
}

func (c *List) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if c.footer.expanded {
				c.footer.expanded = false
				c.footer.cursor = 0
				return c, nil
			}

			if c.input.Value() != "" {
				c.input.SetValue("")
				c.filter.FilterItems("")
				return c, nil
			}

			return c, PopPageCmd
		}
	case QueryChangeMsg:
		if c.OnQueryChange == nil {
			return c, nil
		}

		if string(msg) != c.input.Value() {
			return c, nil
		}

		return c, c.OnQueryChange(string(msg))
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	input, cmd := c.input.Update(msg)

	if input.Value() != c.input.Value() {
		if c.OnQueryChange != nil {
			query := input.Value()
			cmds = append(cmds, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				c.filter.EmptyText = "Loading..."
				if query == c.input.Value() {
					return QueryChangeMsg(query)
				}

				return nil
			}))
		} else {
			c.filter.FilterItems(input.Value())
		}
		if c.filter.Selection() != nil {
			c.footer.SetActions(c.filter.Selection().(ListItem).Actions...)
		} else {
			c.footer.SetActions(c.Actions...)
		}
	}
	c.input = input
	cmds = append(cmds, cmd)

	filter, cmd := c.filter.Update(msg)
	oldSelection := c.filter.Selection()
	newSelection := filter.Selection()
	if newSelection == nil {
		c.footer.SetActions(c.Actions...)
	} else if oldSelection == nil || oldSelection.ID() != newSelection.ID() {
		c.footer.SetActions(newSelection.(ListItem).Actions...)
	}
	c.filter = filter
	cmds = append(cmds, cmd)

	footer, cmd := c.footer.Update(msg)
	c.footer = footer
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c List) View() string {
	inputRow := fmt.Sprintf("   %s", c.input.View())
	return lipgloss.JoinVertical(lipgloss.Left, inputRow, separator(c.width), c.filter.View(), c.footer.View())
}
