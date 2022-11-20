package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
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
	Background    lipgloss.TerminalColor

	choices  []FilterItem
	filtered []FilterItem

	DrawLines bool
	cursor    int
}

func NewFilter() Filter {
	return Filter{}
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
	f.choices = items
}

func (f *Filter) FilterItems(query string) {
	values := make([]string, len(f.choices))
	for i, choice := range f.choices {
		values[i] = choice.FilterValue()
	}
	// If the search field is empty, let's not display the matches
	// (none), but rather display all possible choices.
	if query == "" {
		f.filtered = f.choices
	} else {
		matches := fuzzy.Find(query, values)
		f.filtered = make([]FilterItem, len(matches))
		for i, match := range matches {
			f.filtered[i] = f.choices[match.Index]
		}
	}

	f.Query = query
	// Reset the cursor
	f.cursor = 0
	f.minIndex = 0
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

		if availableHeight > 0 {
			separator := strings.Repeat("─", itemWidth)
			separator = styles.Faint.Render(separator)
			rows = append(rows, separator)
			availableHeight--
		}
	}
	filteredView := lipgloss.JoinVertical(lipgloss.Left, rows...)

	if filteredView == "" {
		var emptyMessage string
		if len(m.choices) == 0 {
			emptyMessage = ""
		} else {
			emptyMessage = "No matches"
		}

		return styles.Faint.Copy().Width(m.Width).Height(m.Height).Padding(0, 3).Render(emptyMessage)
	}

	filteredView = lipgloss.NewStyle().Padding(0, 1).Render(filteredView)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Top, lipgloss.Left, filteredView)
}

func (f Filter) Update(msg tea.Msg) (Filter, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+n", "ctrl+j", "down":
			f.CursorDown()
		case "ctrl+p", "ctrl+k", "up":
			f.CursorUp()
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
	m.cursor = clamp(0, len(m.filtered)-1, m.cursor-1)

	if m.cursor < m.minIndex {
		m.minIndex = m.cursor
	}
}

func (m Filter) nbVisibleItems() int {
	return m.Height / m.itemHeight()
}

func (m *Filter) CursorDown() {
	m.cursor = clamp(0, len(m.filtered)-1, m.cursor+1)
	if m.cursor >= m.minIndex+m.nbVisibleItems() {
		m.minIndex += 1
	}
}

//nolint:unparam
func clamp(min, max, val int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
