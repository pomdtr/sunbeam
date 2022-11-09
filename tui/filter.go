package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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
	viewport *viewport.Model
	textinput.Model

	choices  []FilterItem
	filtered []FilterItem

	DrawLines bool
	cursor    int
}

func NewFilter() *Filter {
	viewport := viewport.New(0, 0)
	viewport.Style = styles.Secondary.Copy().Padding(0, 1)

	ti := textinput.New()
	ti.Focus()
	ti.TextStyle = styles.Secondary
	ti.PlaceholderStyle = styles.Secondary.Copy().Italic(true)
	ti.Prompt = ""
	ti.Placeholder = "Search..."

	return &Filter{viewport: &viewport, Model: ti}
}

func (f *Filter) SetSize(width, height int) {
	f.viewport.Width = width
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
	f.FilterItems(f.Value())
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

	return NewFilterItemChangeCmd(f.Selection())
}

func (m Filter) Init() tea.Cmd { return nil }

func (m Filter) View() string {
	itemViews := make([]string, len(m.filtered))
	for i, item := range m.filtered {
		itemView := item.Render(m.viewport.Width, i == m.cursor)
		if m.DrawLines {
			separator := strings.Repeat("─", m.viewport.Width)
			separator = styles.Secondary.Render(separator)
			itemView = lipgloss.JoinVertical(lipgloss.Left, itemView, separator)
		}
		itemViews[i] = itemView
	}
	filteredView := lipgloss.JoinVertical(lipgloss.Left, itemViews...)
	if filteredView == "" {
		var emptyMessage string
		if len(m.choices) == 0 {
			emptyMessage = ""
		} else {
			emptyMessage = "No matches"
		}
		filteredView = styles.Secondary.Copy().Padding(0, 2).Width(m.viewport.Width).Render(emptyMessage)
	}

	m.viewport.SetContent(filteredView)
	return m.viewport.View()
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
	t, cmd := f.Model.Update(msg)
	cmds = append(cmds, cmd)
	if t.Value() != f.Value() {
		cmd := f.FilterItems(t.Value())
		cmds = append(cmds, cmd)
	}
	f.Model = t

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

	if m.cursor*m.itemHeight() < m.viewport.YOffset {
		m.viewport.SetYOffset(m.cursor * m.itemHeight())
	}
}

func (m *Filter) CursorDown() {
	m.cursor = clamp(0, len(m.filtered)-1, m.cursor+1)
	if m.cursor*m.itemHeight() >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.LineDown(m.itemHeight())
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
