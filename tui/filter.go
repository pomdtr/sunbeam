package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/sahilm/fuzzy"
)

type Filter struct {
	viewport              *viewport.Model
	choices               []string
	matches               []fuzzy.Match
	cursor                int
	indicator             string
	height                int
	matchStyle            lipgloss.Style
	textStyle             lipgloss.Style
	indicatorStyle        lipgloss.Style
	selectedPrefixStyle   lipgloss.Style
	unselectedPrefixStyle lipgloss.Style
	fuzzy                 bool
}

func (f *Filter) Filter(term string) {
	f.matches = fuzzy.Find(term, f.choices)

	// If the search field is empty, let's not display the matches
	// (none), but rather display all possible choices.
	if term == "" {
		f.matches = matchAll(f.choices)
	}

	// It's possible that filtering items have caused fewer matches. So, ensure
	// that the selected index is within the bounds of the number of matches.
	f.cursor = clamp(0, len(f.matches)-1, f.cursor)
}

func (m Filter) Init() tea.Cmd { return nil }
func (m Filter) View() string {
	var s strings.Builder

	for i := range m.matches {
		// For reverse layout, the matches are displayed in reverse order.
		match := m.matches[i]

		// If this is the current selected index, we add a small indicator to
		// represent it. Otherwise, simply pad the string.
		if i == m.cursor {
			s.WriteString(m.indicatorStyle.Render(m.indicator))
		} else {
			s.WriteString(strings.Repeat(" ", runewidth.StringWidth(m.indicator)))
		}

		// For this match, there are a certain number of characters that have
		// caused the match. i.e. fuzzy matching.
		// We should indicate to the users which characters are being matched.
		mi := 0
		var buf strings.Builder
		for ci, c := range match.Str {
			// Check if the current character index matches the current matched
			// index. If so, color the character to indicate a match.
			if mi < len(match.MatchedIndexes) && ci == match.MatchedIndexes[mi] {
				// Flush text buffer.
				s.WriteString(m.textStyle.Render(buf.String()))
				buf.Reset()

				s.WriteString(m.matchStyle.Render(string(c)))
				// We have matched this character, so we never have to check it
				// again. Move on to the next match.
				mi++
			} else {
				// Not a match, buffer a regular character.
				buf.WriteRune(c)
			}
		}
		// Flush text buffer.
		s.WriteString(m.textStyle.Render(buf.String()))

		// We have finished displaying the match with all of it's matched
		// characters highlighted and the rest filled in.
		// Move on to the next match.
		s.WriteRune('\n')
	}

	m.viewport.SetContent(s.String())
	return m.viewport.View()

	// View the input and the filtered choices

	// return m.textinput.View() + "\n" + m.viewport.View()
}

func (m Filter) Update(msg tea.Msg) (Filter, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.height == 0 || m.height > msg.Height {
			m.viewport.Height = msg.Height
		}
		m.viewport.Width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		// case "enter":
		// 	m.textinput.SetValue(m.matches[m.cursor].Str)
		// 	m.textinput.CursorEnd()
		case "ctrl+n", "ctrl+j", "down":
			m.CursorDown()
		case "ctrl+p", "ctrl+k", "up":
			m.CursorUp()
		}
	}

	return m, cmd
}

func (m *Filter) CursorUp() {
	m.cursor = clamp(0, len(m.matches)-1, m.cursor-1)
	if m.cursor < m.viewport.YOffset {
		m.viewport.SetYOffset(m.cursor)
	}
}

func (m *Filter) CursorDown() {
	m.cursor = clamp(0, len(m.matches)-1, m.cursor+1)
	if m.cursor >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.LineDown(1)
	}
}

func matchAll(options []string) []fuzzy.Match {
	matches := make([]fuzzy.Match, len(options))
	for i, option := range options {
		matches[i] = fuzzy.Match{Str: option}
	}
	return matches
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
