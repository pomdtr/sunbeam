package internal

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/internal/fzf"
	"github.com/pomdtr/sunbeam/utils"
)

type FilterItem interface {
	FilterValue() string
	Render(width int, selected bool) string
	ID() string
}

type Filter struct {
	minIndex      int
	Width, Height int
	Query         string
	Less          func(i, j FilterItem) bool

	items     []FilterItem
	filtered  []FilterItem
	EmptyText string

	DrawLines bool
	cursor    int
}

func NewFilter() Filter {
	return Filter{
		EmptyText: "No results",
	}
}

func (f *Filter) SetSize(width, height int) {
	f.Width = width
	f.Height = height
}

func (f Filter) Selection() FilterItem {
	if f.cursor >= len(f.filtered) || f.cursor < 0 {
		return nil
	}
	return f.filtered[f.cursor]
}

func (f *Filter) SetItems(items []FilterItem) {
	f.items = items
	f.filtered = items
}

type Match struct {
	item  *FilterItem
	score int
}

func (f *Filter) FilterItems(query string) {
	f.Query = query
	values := make([]string, len(f.items))
	for i, choice := range f.items {
		values[i] = choice.FilterValue()
	}
	// If the search field is empty, let's not display the matches
	// (none), but rather display all possible choices.
	var filtered []FilterItem
	if query == "" {
		filtered = f.items
	} else {
		matches := make([]Match, 0)
		for i := 0; i < len(f.items); i++ {
			score := fzf.Score(f.items[i].FilterValue(), query)
			if score > 0 {
				matches = append(matches, Match{&f.items[i], score})
			}
		}

		sort.SliceStable(matches, func(i, j int) bool {
			return matches[i].score > matches[j].score
		})

		filtered = make([]FilterItem, len(matches))
		for i, match := range matches {
			filtered[i] = *match.item
		}
	}

	if f.Less != nil {
		sort.SliceStable(filtered, func(i, j int) bool {
			return f.Less(filtered[i], filtered[j])
		})
	}

	f.filtered = filtered

	// Reset the cursor
	f.cursor = 0
	f.minIndex = 0
}

func (f *Filter) Select(id string) {
	for i, item := range f.filtered {
		if item.ID() == id {
			f.cursor = i
		}
	}
}

func (m Filter) Init() tea.Cmd { return nil }

func (m Filter) View() string {
	itemWidth := m.Width - 2
	rows := make([]string, 0)
	index := m.minIndex
	availableHeight := m.Height
	for availableHeight > 0 && index < len(m.filtered) {
		item := m.filtered[index]
		itemView := item.Render(itemWidth, index == m.cursor)
		rows = append(rows, itemView)

		index++
		availableHeight--

		if availableHeight > 0 && m.DrawLines {
			separator := strings.Repeat("─", itemWidth)
			separator = lipgloss.NewStyle().Faint(true).Render(separator)
			rows = append(rows, separator)
			availableHeight--
		}
	}

	if len(rows) == 0 {
		return lipgloss.NewStyle().Faint(true).Copy().Width(m.Width).Height(m.Height).Padding(0, 3).Render(m.EmptyText)
	}

	filteredView := lipgloss.JoinVertical(lipgloss.Left, rows...)
	filteredView = lipgloss.NewStyle().Padding(0, 1).Render(filteredView)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Top, lipgloss.Left, filteredView)
}

func (f Filter) Update(msg tea.Msg) (Filter, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down", "ctrl+j":
			f.CursorDown()
		case "shift+tab", "up", "ctrl+k":
			f.CursorUp()
		case "ctrl+u":
			shift := utils.Min(f.nbVisibleItems(), f.cursor)
			for i := 0; i < shift; i++ {
				f.CursorUp()
			}
		case "ctrl+d":
			shift := utils.Min(f.nbVisibleItems(), len(f.filtered)-f.cursor-1)
			for i := 0; i < shift; i++ {
				f.CursorDown()
			}
		}
	}

	return f, nil
}

func (m Filter) itemHeight() int {
	if m.DrawLines {
		return 2
	}
	return 1
}

func (m *Filter) CursorUp() {
	if m.cursor > 0 {
		m.cursor = m.cursor - 1

		if m.cursor < m.minIndex {
			m.minIndex = m.cursor
		}
	} else {
		m.cursor = len(m.filtered) - 1
		m.minIndex = utils.Max(0, m.cursor-m.nbVisibleItems()+1)
	}
}

func (m Filter) nbVisibleItems() int {
	return m.Height/m.itemHeight() + m.Height%m.itemHeight()
}

func (m *Filter) CursorDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor += 1
		if m.cursor >= m.minIndex+m.nbVisibleItems() {
			m.minIndex += 1
		}
	} else {
		m.cursor = 0
		m.minIndex = 0
	}
}
