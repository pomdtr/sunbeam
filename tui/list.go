package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/sahilm/fuzzy"
)

type ListContainer struct {
	width, height             int
	startIndex, selectedIndex int

	title         string
	textInput     *textinput.Model
	filteredItems []api.ListItem
	items         []api.ListItem
	runAction     func(api.ScriptAction) tea.Cmd
}

func NewListContainer(title string, items []api.ListItem, runAction func(api.ScriptAction) tea.Cmd) Container {
	t := textinput.New()
	t.Prompt = "> "
	t.Placeholder = "Search..."

	return &ListContainer{
		textInput: &t,
		title:     title,

		items:         items,
		runAction:     runAction,
		filteredItems: items,
	}
}

// Rank defines a rank for a given item.
type Rank struct {
	// The index of the item in the original input.
	Index int
	// Indices of the actual word that were matched against the filter term.
	MatchedIndexes []int
}

// filterItems uses the sahilm/fuzzy to filter through the list.
// This is set by default.
func filterItems(term string, items []api.ListItem) []api.ListItem {
	targets := make([]string, len(items))
	for i, item := range items {
		targets[i] = item.Title
	}
	var ranks = fuzzy.Find(term, targets)
	sort.Stable(ranks)
	filteredItems := make([]api.ListItem, len(ranks))
	for i, r := range ranks {
		filteredItems[i] = items[r.Index]
	}
	return filteredItems
}

func (c *ListContainer) Init() tea.Cmd {
	return c.textInput.Focus()
}

func (c *ListContainer) SetSize(width, height int) {
	c.width, c.height = width, height-lipgloss.Height(c.headerView())-lipgloss.Height(c.footerView())-1
	c.updateIndexes(c.selectedIndex)
}

func (c ListContainer) SelectedItem() (api.ListItem, bool) {
	if c.selectedIndex >= len(c.filteredItems) {
		return api.ListItem{}, false
	}
	return c.filteredItems[c.selectedIndex], true
}

func (c *ListContainer) headerView() string {
	input := c.textInput.View()
	line := strings.Repeat("─", c.width)
	return lipgloss.JoinVertical(lipgloss.Left, input, line)
}

func (c *ListContainer) footerView() string {
	selectedItem, ok := c.SelectedItem()
	if !ok {
		return SunbeamFooter(c.width, c.title)
	}

	if len(selectedItem.Actions) > 0 {
		return SunbeamFooterWithActions(c.width, c.title, selectedItem.Actions[0].Title)
	} else {
		return SunbeamFooter(c.width, c.title)
	}
}

func (c *ListContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	var cmds []tea.Cmd

	selectedItem, _ := c.SelectedItem()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return c, c.runAction(selectedItem.Actions[0])
		case tea.KeyDown, tea.KeyTab, tea.KeyCtrlJ:
			if c.selectedIndex < len(c.filteredItems)-1 {
				c.updateIndexes(c.selectedIndex + 1)
			}
		case tea.KeyUp, tea.KeyShiftTab, tea.KeyCtrlK:
			if c.selectedIndex > 0 {
				c.updateIndexes(c.selectedIndex - 1)
			}
		case tea.KeyEscape:
			if c.textInput.Value() != "" {
				c.textInput.SetValue("")
			} else {
				return c, PopCmd
			}
		default:
			for _, action := range selectedItem.Actions {
				if action.Keybind == msg.String() {
					return c, c.runAction(action)
				}
			}
		}
	}

	t, cmd := c.textInput.Update(msg)
	cmds = append(cmds, cmd)
	if t.Value() != c.textInput.Value() {
		if t.Value() == "" {
			c.filteredItems = c.items
		} else {
			c.filteredItems = filterItems(t.Value(), c.items)
		}
		c.updateIndexes(0)
	}
	c.textInput = &t

	return c, tea.Batch(cmds...)
}

func (c *ListContainer) updateIndexes(selectedIndex int) {
	for selectedIndex < c.startIndex {
		c.startIndex--
	}
	for selectedIndex > c.startIndex+c.height {
		c.startIndex++
	}
	c.selectedIndex = selectedIndex
}

func (c *ListContainer) View() string {
	rows := make([]string, 0)
	items := c.filteredItems

	endIndex := utils.Min(c.startIndex+c.height+1, len(items))
	for i := c.startIndex; i < endIndex; i++ {
		if i == c.selectedIndex {
			rows = append(rows, fmt.Sprintf("> %s - %s", items[i].Title, items[i].Subtitle))
		} else {
			rows = append(rows, fmt.Sprintf("  %s - %s", items[i].Title, items[i].Subtitle))
		}
	}

	for i := len(rows) - 1; i < c.height; i++ {
		rows = append(rows, "")
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), strings.Join(rows, "\n"), c.footerView())
}

var NewErrorCmd = func(format string, values ...any) func() tea.Msg {
	return func() tea.Msg {
		return fmt.Errorf(format, values...)
	}
}
