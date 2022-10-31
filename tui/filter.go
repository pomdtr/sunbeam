package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

type FilterItem interface {
	FilterValue() string
	Render(width int, selected bool) string
}

type Filter struct {
	viewport   *viewport.Model
	choices    []FilterItem
	filtered   []FilterItem
	itemHeight int

	cursor int
	height int
}

func (f *Filter) SetSize(width, height int) {
	f.viewport.Width = width - 2
	f.viewport.Height = height
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

func (f *Filter) Filter(query string) {
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

	// Reset the cursor
	f.cursor = 0
}

func (m Filter) Init() tea.Cmd { return nil }

func (m Filter) View() string {
	itemsView := make([]string, len(m.filtered))
	for i, item := range m.filtered {
		itemsView[i] = item.Render(m.viewport.Width, i == m.cursor)
	}
	filteredView := lipgloss.JoinVertical(lipgloss.Left, itemsView...)

	m.viewport.SetContent(filteredView)
	return lipgloss.NewStyle().Padding(0, 1).Render(m.viewport.View())
}

func (m Filter) Update(msg tea.Msg) (Filter, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+n", "ctrl+j", "down":
			m.CursorDown()
		case "ctrl+p", "ctrl+k", "up":
			m.CursorUp()
		}
	}

	return m, cmd
}

func (m *Filter) CursorUp() {
	m.cursor = clamp(0, len(m.filtered)-1, m.cursor-1)
	if m.cursor*m.itemHeight < m.viewport.YOffset {
		m.viewport.SetYOffset(m.cursor * m.itemHeight)
	}
}

func (m *Filter) CursorDown() {
	m.cursor = clamp(0, len(m.filtered)-1, m.cursor+1)
	if m.cursor*m.itemHeight >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.LineDown(m.itemHeight)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
