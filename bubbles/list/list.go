// Package list provides a feature-rich Bubble Tea component for browsing
// a general purpose list of items.
package list

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// Item is an item that appears in the list.
type Item interface {
	// Filter value is the value we use when filtering against this item when
	// we're filtering the list.
	FilterValue() string
}

// ItemDelegate encapsulates the general functionality for all list items. The
// benefit to separating this logic from the item itself is that you can change
// the functionality of items without changing the actual items themselves.
//
// Note that if the delegate also implements help.KeyMap delegate-related
// help items will be added to the help view.
type ItemDelegate interface {
	// Render renders the item's view.
	Render(w io.Writer, m Model, index int, item Item)

	// Height is the height of the list item.
	Height() int

	// Spacing is the size of the horizontal gap between list items in cells.
	Spacing() int

	// Update is the update loop for items. All messages in the list's update
	// loop will pass through here except when the user is setting a filter.
	// Use this method to perform item-level updates appropriate to this
	// delegate.
	Update(msg tea.Msg, m *Model) tea.Cmd
}

type filteredItem struct {
	item    Item  // item matched
	matches []int // rune indices of matched items
}

type filteredItems []filteredItem

func (f filteredItems) items() []Item {
	agg := make([]Item, len(f))
	for i, v := range f {
		agg[i] = v.item
	}
	return agg
}

// FilterMatchesMsg contains data about items matched during filtering. The
// message should be routed to Update for processing.
type FilterMatchesMsg []filteredItem

// FilterFunc takes a term and a list of strings to search through
// (defined by Item#FilterValue).
// It should return a sorted list of ranks.
type FilterFunc func(string, []string) []Rank

// Rank defines a rank for a given item.
type Rank struct {
	// The index of the item in the original input.
	Index int
	// Indices of the actual word that were matched against the filter term.
	MatchedIndexes []int
}

// DefaultFilter uses the sahilm/fuzzy to filter through the list.
// This is set by default.
func DefaultFilter(term string, targets []string) []Rank {
	var ranks = fuzzy.Find(term, targets)
	sort.Stable(ranks)
	result := make([]Rank, len(ranks))
	for i, r := range ranks {
		result[i] = Rank{
			Index:          r.Index,
			MatchedIndexes: r.MatchedIndexes,
		}
	}
	return result
}

type statusMessageTimeoutMsg struct{}

// Model contains the state of this component.
type Model struct {
	itemNameSingular string
	itemNamePlural   string

	Styles Styles

	// Key mappings for navigating the list.
	KeyMap KeyMap

	// Filter is used to filter the list.
	Filter FilterFunc

	disableQuitKeybindings bool

	width       int
	height      int
	Paginator   paginator.Model
	cursor      int
	FilterInput textinput.Model

	// The master set of items we're working with.
	items []Item

	// Filtered items we're currently displaying. Filtering, toggles and so on
	// will alter this slice so we can show what is relevant. For that reason,
	// this field should be considered ephemeral.
	filteredItems filteredItems

	delegate ItemDelegate
}

// New returns a new model with sensible defaults.
func New(items []Item, delegate ItemDelegate, width, height int) Model {
	styles := DefaultStyles()

	// TODO: remove this when we have a better way to handle filtering
	filterInput := textinput.NewModel()
	filterInput.Prompt = "Filter: "
	filterInput.PromptStyle = styles.FilterPrompt
	filterInput.CursorStyle = styles.FilterCursor
	filterInput.CharLimit = 64
	filterInput.Focus()

	p := paginator.NewModel()
	p.Type = paginator.Dots

	m := Model{
		itemNameSingular: "item",
		itemNamePlural:   "items",
		KeyMap:           DefaultKeyMap(),
		Filter:           DefaultFilter,
		Styles:           styles,
		FilterInput:      filterInput,

		width:     width,
		height:    height,
		delegate:  delegate,
		items:     items,
		Paginator: p,
	}

	m.updatePagination()
	return m
}

// Items returns the items in the list.
func (m Model) Items() []Item {
	return m.items
}

// Set the items available in the list. This returns a command.
func (m *Model) SetItems(i []Item) tea.Cmd {
	var cmd tea.Cmd
	m.items = i

	m.filteredItems = nil
	cmd = filterItems(*m)

	m.updatePagination()
	return cmd
}

// Set the item delegate.
func (m *Model) SetDelegate(d ItemDelegate) {
	m.delegate = d
	m.updatePagination()
}

// VisibleItems returns the total items available to be shown.
func (m Model) VisibleItems() []Item {
	return m.filteredItems.items()
}

// SelectedItems returns the current selected item in the list.
func (m Model) SelectedItem() Item {
	i := m.Index()

	items := m.VisibleItems()
	if i < 0 || len(items) == 0 || len(items) <= i {
		return nil
	}

	return items[i]
}

// Index returns the index of the currently selected item as it appears in the
// entire slice of items.
func (m Model) Index() int {
	return m.Paginator.Page*m.Paginator.PerPage + m.cursor
}

// Cursor returns the index of the cursor on the current page.
func (m Model) Cursor() int {
	return m.cursor
}

// CursorUp moves the cursor up. This can also move the state to the previous
// page.
func (m *Model) CursorUp() {
	m.cursor--

	// If we're at the start, stop
	if m.cursor < 0 && m.Paginator.Page == 0 {
		m.cursor = 0
		return
	}

	// Move the cursor as normal
	if m.cursor >= 0 {
		return
	}

	// Go to the previous page
	m.Paginator.PrevPage()
	m.cursor = m.Paginator.ItemsOnPage(len(m.VisibleItems())) - 1
}

// CursorDown moves the cursor down. This can also advance the state to the
// next page.
func (m *Model) CursorDown() {
	itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))

	m.cursor++

	// If we're at the end, stop
	if m.cursor < itemsOnPage {
		return
	}

	// Go to the next page
	if !m.Paginator.OnLastPage() {
		m.Paginator.NextPage()
		m.cursor = 0
		return
	}

	// During filtering the cursor position can exceed the number of
	// itemsOnPage. It's more intuitive to start the cursor at the
	// topmost position when moving it down in this scenario.
	if m.cursor > itemsOnPage {
		m.cursor = 0
		return
	}

	m.cursor = itemsOnPage - 1
}

// PrevPage moves to the previous page, if available.
func (m Model) PrevPage() {
	m.Paginator.PrevPage()
}

// NextPage moves to the next page, if available.
func (m Model) NextPage() {
	m.Paginator.NextPage()
}

// FilterValue returns the current value of the filter.
func (m Model) FilterValue() string {
	return m.FilterInput.Value()
}

// Width returns the current width setting.
func (m Model) Width() int {
	return m.width
}

// Height returns the current height setting.
func (m Model) Height() int {
	return m.height
}

// SetSize sets the width and height of this component.
func (m *Model) SetSize(width, height int) {
	m.setSize(width, height)
}

// SetWidth sets the width of this component.
func (m *Model) SetWidth(v int) {
	m.setSize(v, m.height)
}

// SetHeight sets the height of this component.
func (m *Model) SetHeight(v int) {
	m.setSize(m.width, v)
}

func (m *Model) setSize(width, height int) {

	m.width = width
	m.height = height
	m.updatePagination()
}

func (m Model) itemsAsFilterItems() filteredItems {
	fi := make([]filteredItem, len(m.items))
	for i, item := range m.items {
		fi[i] = filteredItem{
			item: item,
		}
	}
	return filteredItems(fi)
}

// Update pagination according to the amount of items for the current state.
func (m *Model) updatePagination() {
	index := m.Index()
	availHeight := m.height

	m.Paginator.PerPage = max(1, availHeight/(m.delegate.Height()+m.delegate.Spacing()))

	if pages := len(m.VisibleItems()); pages < 1 {
		m.Paginator.SetTotalPages(1)
	} else {
		m.Paginator.SetTotalPages(pages)
	}

	// Restore index
	m.Paginator.Page = index / m.Paginator.PerPage
	m.cursor = index % m.Paginator.PerPage

	// Make sure the page stays in bounds
	if m.Paginator.Page >= m.Paginator.TotalPages-1 {
		m.Paginator.Page = max(0, m.Paginator.TotalPages-1)
	}
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case FilterMatchesMsg:
		m.filteredItems = filteredItems(msg)
		return m, nil
	}

	cmds = append(cmds, m.handleFiltering(msg))
	cmds = append(cmds, m.handleBrowsing(msg))

	return m, tea.Batch(cmds...)
}

func (m *Model) enableLiveFiltering() tea.Cmd {
	m.Paginator.Page = 0
	m.cursor = 0
	m.FilterInput.CursorEnd()
	m.FilterInput.Focus()
	return tea.Batch(textinput.Blink, filterItems(*m))
}

// Updates for when a user is browsing the list.
func (m *Model) handleBrowsing(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	numItems := len(m.VisibleItems())

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.CursorUp):
			m.CursorUp()

		case key.Matches(msg, m.KeyMap.CursorDown):
			m.CursorDown()

		case key.Matches(msg, m.KeyMap.PrevPage):
			m.Paginator.PrevPage()

		case key.Matches(msg, m.KeyMap.NextPage):
			m.Paginator.NextPage()

		case key.Matches(msg, m.KeyMap.GoToStart):
			m.Paginator.Page = 0
			m.cursor = 0

		case key.Matches(msg, m.KeyMap.GoToEnd):
			m.Paginator.Page = m.Paginator.TotalPages - 1
			m.cursor = m.Paginator.ItemsOnPage(numItems) - 1
		}
	}

	cmd := m.delegate.Update(msg, m)
	cmds = append(cmds, cmd)

	// Keep the index in bounds when paginating
	itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))
	if m.cursor > itemsOnPage-1 {
		m.cursor = max(0, itemsOnPage-1)
	}

	return tea.Batch(cmds...)
}

// Updates for when a user is in the filter editing interface.
func (m *Model) handleFiltering(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Update the filter text input component
	newFilterInputModel, inputCmd := m.FilterInput.Update(msg)
	filterChanged := m.FilterInput.Value() != newFilterInputModel.Value()
	m.FilterInput = newFilterInputModel
	cmds = append(cmds, inputCmd)

	// If the filtering input has changed, request updated filtering
	if filterChanged {
		cmds = append(cmds, filterItems(*m))
	}

	// Update pagination
	m.updatePagination()

	return tea.Batch(cmds...)
}

// View renders the component.
func (m Model) View() string {
	return lipgloss.NewStyle().Height(m.height).Render(m.populatedView())
}

func (m Model) populatedView() string {
	items := m.VisibleItems()

	var b strings.Builder

	// Empty states
	if len(items) == 0 {
		return m.Styles.NoItems.Render("No " + m.itemNamePlural + " found.")
	}

	if len(items) > 0 {
		start, end := m.Paginator.GetSliceBounds(len(items))
		docs := items[start:end]

		for i, item := range docs {
			m.delegate.Render(&b, m, i+start, item)
			if i != len(docs)-1 {
				fmt.Fprint(&b, strings.Repeat("\n", m.delegate.Spacing()+1))
			}
		}
	}

	// If there aren't enough items to fill up this page (always the last page)
	// then we need to add some newlines to fill up the space where items would
	// have been.
	itemsOnPage := m.Paginator.ItemsOnPage(len(items))
	if itemsOnPage < m.Paginator.PerPage {
		n := (m.Paginator.PerPage - itemsOnPage) * (m.delegate.Height() + m.delegate.Spacing())
		if len(items) == 0 {
			n -= m.delegate.Height() - 1
		}
		fmt.Fprint(&b, strings.Repeat("\n", n))
	}

	return b.String()
}

func filterItems(m Model) tea.Cmd {
	return func() tea.Msg {
		if m.FilterInput.Value() == "" {
			return FilterMatchesMsg(m.itemsAsFilterItems()) // return nothing
		}

		targets := []string{}
		items := m.items

		for _, t := range items {
			targets = append(targets, t.FilterValue())
		}

		filterMatches := []filteredItem{}
		for _, r := range m.Filter(m.FilterInput.Value(), targets) {
			filterMatches = append(filterMatches, filteredItem{
				item:    items[r.Index],
				matches: r.MatchedIndexes,
			})
		}

		return FilterMatchesMsg(filterMatches)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
