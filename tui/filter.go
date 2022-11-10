package tui

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/sahilm/fuzzy"
)

type FilterItem interface {
	FilterValue() string
	Render(width int, selected bool) string
}

type Filter struct {
	TextInput     textinput.Model
	minIndex      int
	Width, Height int

	choices  []FilterItem
	filtered []FilterItem

	DrawLines bool
	cursor    int
}

func NewFilter() *Filter {
	ti := textinput.New()
	ti.TextStyle = styles.Text
	ti.PlaceholderStyle = styles.Text.Copy().Italic(true)
	ti.Prompt = ""
	ti.Placeholder = "Search..."

	return &Filter{TextInput: ti}
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
	f.FilterItems(f.TextInput.Value())
}

func (f *Filter) FilterItems(term string) tea.Cmd {
	values := make([]string, len(f.choices))
	for i, choice := range f.choices {
		values[i] = choice.FilterValue()
	}
	// If the search field is empty, let's not display the matches
	// (none), but rather display all possible choices.
	if term == "" {
		f.filtered = f.choices
	} else {
		matches := fuzzy.Find(term, values)
		f.filtered = make([]FilterItem, len(matches))
		for i, match := range matches {
			f.filtered[i] = f.choices[match.Index]
		}
	}

	// Reset the cursor
	f.cursor = 0
	f.minIndex = 0

	return NewFilterItemChangeCmd(f.Selection())
}

func (m Filter) Init() tea.Cmd { return nil }

func (m Filter) View() string {
	maxIndex := utils.Min(m.minIndex+m.nbVisibleItems(), len(m.filtered))
	log.Println(m.minIndex, maxIndex, maxIndex-m.minIndex)
	itemViews := make([]string, maxIndex-m.minIndex)
	for i := m.minIndex; i < maxIndex; i++ {
		log.Println(i)
		item := m.filtered[i]
		itemView := item.Render(m.Width, i == m.cursor)
		if m.DrawLines {
			separator := strings.Repeat("─", m.Width)
			separator = styles.Text.Render(separator)
			itemView = lipgloss.JoinVertical(lipgloss.Left, itemView, separator)
		}
		itemViews[i-m.minIndex] = itemView
	}
	filteredView := lipgloss.JoinVertical(lipgloss.Left, itemViews...)

	if filteredView == "" {
		var emptyMessage string
		if len(m.choices) == 0 {
			emptyMessage = ""
		} else {
			emptyMessage = "No matches"
		}
		filteredView = styles.Text.Copy().Padding(0, 2).Width(m.Width).Render(emptyMessage)
	}

	return lipgloss.Place(m.Width, m.Height, lipgloss.Top, lipgloss.Left, filteredView, lipgloss.WithWhitespaceBackground(theme.Bg()))
}

func (f Filter) Update(msg tea.Msg) (*Filter, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+n", "ctrl+j", "down":
			f.CursorDown()
			return &f, NewFilterItemChangeCmd(f.Selection())
		case "ctrl+p", "ctrl+k", "up":
			f.CursorUp()
			return &f, NewFilterItemChangeCmd(f.Selection())
		}
	}

	var cmds []tea.Cmd
	t, cmd := f.TextInput.Update(msg)
	cmds = append(cmds, cmd)
	if t.Value() != f.TextInput.Value() {
		cmd := f.FilterItems(t.Value())
		cmds = append(cmds, cmd)
	}
	f.TextInput = t

	return &f, tea.Batch(cmds...)
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

type FilterItemChange struct {
	FilterItem FilterItem
}

func NewFilterItemChangeCmd(filterItem FilterItem) tea.Cmd {
	return func() tea.Msg {
		return FilterItemChange{FilterItem: filterItem}
	}
}
