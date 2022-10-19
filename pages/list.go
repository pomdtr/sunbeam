package pages

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/bubbles"
	"github.com/pomdtr/sunbeam/scripts"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/sahilm/fuzzy"
)

type ListContainer struct {
	textInput     *textinput.Model
	filteredItems []scripts.ScriptItem
	selectedIdx   int
	startIndex    int
	width         int
	height        int
	response      *scripts.ListResponse
}

func NewListContainer(res *scripts.ListResponse) Container {
	t := textinput.New()
	t.Prompt = "> "
	t.Placeholder = res.SearchBarPlaceholder
	if res.SearchBarPlaceholder != "" {
		t.Placeholder = res.SearchBarPlaceholder
	} else {
		t.Placeholder = "Search..."
	}

	return &ListContainer{
		textInput:     &t,
		filteredItems: res.Items,
		response:      res,
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
func filterItems(term string, items []scripts.ScriptItem) []scripts.ScriptItem {
	targets := make([]string, len(items))
	for i, item := range items {
		targets[i] = item.Title
	}
	var ranks = fuzzy.Find(term, targets)
	sort.Stable(ranks)
	filteredItems := make([]scripts.ScriptItem, len(ranks))
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
	c.updateIndexes(c.selectedIdx)
}

func (c ListContainer) SelectedItem() (scripts.ScriptItem, bool) {
	if c.selectedIdx >= len(c.filteredItems) {
		return scripts.ScriptItem{}, false
	}
	return c.filteredItems[c.selectedIdx], true
}

func (c *ListContainer) headerView() string {
	input := c.textInput.View()
	line := strings.Repeat("─", c.width)
	return lipgloss.JoinVertical(lipgloss.Left, input, line)
}

func (c *ListContainer) footerView() string {
	selectedItem, ok := c.SelectedItem()
	if !ok {
		return bubbles.SunbeamFooter(c.width, c.response.Title)
	}

	if len(selectedItem.Actions) > 0 {
		return bubbles.SunbeamFooterWithActions(c.width, c.response.Title, selectedItem.Actions[0].Title)
	} else {
		return bubbles.SunbeamFooter(c.width, c.response.Title)
	}
}

func (c *ListContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	var cmds []tea.Cmd

	selectedItem, _ := c.SelectedItem()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return c, utils.SendMsg(selectedItem.Actions[0])
		case tea.KeyDown, tea.KeyTab, tea.KeyCtrlJ:
			if c.selectedIdx < len(c.response.Items)-1 {
				c.updateIndexes(c.selectedIdx + 1)
			}
		case tea.KeyUp, tea.KeyShiftTab, tea.KeyCtrlK:
			if c.selectedIdx > 0 {
				c.updateIndexes(c.selectedIdx - 1)
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
					return c, utils.SendMsg(action)
				}
			}
		}
	case scripts.ListResponse:
		c.response = &msg
		c.selectedIdx = 0
		return c, nil
	}

	t, cmd := c.textInput.Update(msg)
	cmds = append(cmds, cmd)
	if t.Value() == "" {
		c.filteredItems = c.response.Items
	} else if t.Value() != c.textInput.Value() {
		c.filteredItems = filterItems(t.Value(), c.response.Items)
		c.selectedIdx = 0
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
	c.selectedIdx = selectedIndex
}

func (c *ListContainer) View() string {
	rows := make([]string, 0)
	items := c.filteredItems

	endIndex := utils.Min(c.startIndex+c.height+1, len(items))
	for i := c.startIndex; i < endIndex; i++ {
		if i == c.selectedIdx {
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
