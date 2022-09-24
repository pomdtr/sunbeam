package pages

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/pomdtr/sunbeam/bubbles/list"
)

var (
	NormalTitle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	NormalDesc = NormalTitle.Copy().
			Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})

	SelectedTitle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"})

	SelectedDesc = SelectedTitle.Copy().
			Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})

	FilterMatch = lipgloss.NewStyle().Underline(true)
)

type ItemDelegate struct {
	list.ItemDelegate
}

func NewItemDelegate() ItemDelegate {
	return ItemDelegate{
		ItemDelegate: list.NewDefaultDelegate(),
	}
}

// Height is the height of the list item.
func (i ItemDelegate) Height() int {
	return 1
}

// Spacing is the size of the horizontal gap between list items in cells.
func (i ItemDelegate) Spacing() int {
	return 1
}

// Render prints an item.
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var (
		title, desc string
	)

	if i, ok := item.(list.DefaultItem); ok {
		title = i.Title()
		desc = i.Description()
	} else {
		return
	}

	if m.Width() <= 0 {
		// short-circuit
		return
	}

	// Prevent text from exceeding list width
	textwidth := uint(m.Width() - NormalTitle.GetPaddingLeft() - NormalTitle.GetPaddingRight())
	title = truncate.StringWithTail(title, textwidth, "…")

	// Conditions
	var (
		isSelected = index == m.Index()
	)

	if isSelected {
		title = SelectedTitle.Render(title)
		desc = SelectedDesc.Render(desc)
	} else {
		title = NormalTitle.Render(title)
		desc = NormalDesc.Render(desc)
	}

	fmt.Fprintf(w, "%s %s", title, desc)
}
