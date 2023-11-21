package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/internal/types"
)

type List struct {
	width, height int

	query string
	input textinput.Model

	spinner   spinner.Model
	filter    Filter
	statusBar StatusBar

	focus         ListFocus
	isLoading     bool
	Actions       []types.Action
	OnQueryChange func(string) tea.Cmd
	OnSelect      func(string) tea.Cmd
}

type ListFocus string

var (
	ListFocusItems   ListFocus = "items"
	ListFocusActions ListFocus = "actions"
)

type QueryChangeMsg string

func NewList(items ...types.ListItem) *List {
	filter := NewFilter()
	filter.DrawLines = true

	statusBar := NewStatusBar()

	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Search Items..."

	list := &List{
		spinner:   spinner.New(),
		input:     input,
		filter:    filter,
		statusBar: statusBar,
		focus:     ListFocusItems,
	}

	list.SetItems(items...)
	return list
}

func (l *List) ResetSelection() {
	l.filter.ResetSelection()

	if selection := l.filter.Selection(); selection != nil {
		l.statusBar.SetActions(selection.(ListItem).Actions...)
	} else {
		l.statusBar.SetActions(l.Actions...)
	}
}

func (l *List) SetActions(actions ...types.Action) {
	l.Actions = actions
	if l.filter.Selection() == nil {
		l.statusBar.SetActions(actions...)
	}
}

func (l *List) SetEmptyText(text string) {
	l.filter.EmptyText = text
}

func (c *List) Init() tea.Cmd {
	if c.focus == ListFocusActions {
		c.focus = ListFocusItems
		c.input.Placeholder = "Search Items..."
		c.input.SetValue(c.query)
	}
	return c.input.Focus()
}

func (c *List) Focus() tea.Cmd {
	return c.input.Focus()
}

func (c *List) Blur() tea.Cmd {
	return nil
}

func (c *List) SetQuery(query string) tea.Cmd {
	c.input.SetValue(query)

	if c.focus == ListFocusItems {
		c.query = query
		if c.OnQueryChange != nil {
			return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				c.filter.EmptyText = "Loading..."
				if query == c.input.Value() {
					return QueryChangeMsg(query)
				}

				return nil
			})
		}

		c.FilterItems(query)
		c.filter.ResetSelection()
	} else {
		c.statusBar.FilterActions(query)
	}

	return nil
}

func (c *List) FilterItems(query string) {
	c.filter.FilterItems(query)
	selection := c.filter.Selection()
	if selection == nil {
		c.statusBar.SetActions(c.Actions...)
	} else {
		c.statusBar.SetActions(selection.(ListItem).Actions...)
	}
}

func (c *List) SetSize(width, height int) {
	c.width, c.height = width, height
	availableHeight := max(0, height-4)

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
	if c.OnQueryChange == nil {
		c.FilterItems(c.Query())
	}
}

func (c *List) SetIsLoading(isLoading bool) tea.Cmd {
	c.isLoading = isLoading
	if isLoading {
		return c.spinner.Tick
	}
	return nil
}

func (c List) Query() string {
	return c.query
}

func (c *List) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if c.input.Value() != "" {
				return c, c.SetQuery("")
			}

			if c.statusBar.expanded {
				c.statusBar.expanded = false
				c.statusBar.cursor = 0
				c.focus = ListFocusItems
				c.input.Placeholder = "Search Items..."
				c.input.SetValue(c.query)
				return c, nil
			}

			return c, PopPageCmd
		case "tab":
			if c.statusBar.expanded {
				break
			}

			c.input.SetValue("")
			c.input.Placeholder = "Search Actions..."
			c.statusBar.expanded = true
			c.focus = ListFocusActions
			return c, nil
		case "right", "left":
			if c.statusBar.expanded {
				statusBar, cmd := c.statusBar.Update(msg)
				c.statusBar = statusBar
				return c, cmd
			}

			input, cmd := c.input.Update(msg)
			c.input = input
			return c, cmd
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

	statusBar, cmd := c.statusBar.Update(msg)
	c.statusBar = statusBar
	if cmd != nil {
		return c, cmd
	}

	input, cmd := c.input.Update(msg)
	if input.Value() != c.input.Value() {
		cmds = append(cmds, c.SetQuery(input.Value()))
	}

	c.input = input
	cmds = append(cmds, cmd)

	filter, cmd := c.filter.Update(msg)
	oldSelection := c.filter.Selection()
	newSelection := filter.Selection()
	if newSelection == nil {
		c.statusBar.SetActions(c.Actions...)
	} else if oldSelection == nil || oldSelection.ID() != newSelection.ID() {
		c.statusBar.SetActions(newSelection.(ListItem).Actions...)
	}
	c.filter = filter
	cmds = append(cmds, cmd)

	if c.isLoading {
		c.spinner, cmd = c.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

func (c List) View() string {
	var headerRow string
	if c.isLoading {
		headerRow = fmt.Sprintf(" %s %s", c.spinner.View(), c.input.View())
	} else {
		headerRow = fmt.Sprintf("   %s", c.input.View())
	}
	return lipgloss.JoinVertical(lipgloss.Left, headerRow, separator(c.width), c.filter.View(), c.statusBar.View())
}
