package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/pomdtr/sunbeam/internal/types"
	"github.com/pomdtr/sunbeam/internal/utils"
)

type List struct {
	width, height int

	query string
	input textinput.Model

	spinner   spinner.Model
	filter    Filter
	viewport  viewport.Model
	statusBar StatusBar

	showDetail bool
	isLoading  bool

	focus         ListFocus
	Actions       []types.Action
	OnQueryChange func(string) tea.Cmd
	OnSelect      func(string) tea.Cmd
}

type ListFocus string

var (
	ListFocusItems   ListFocus = "items"
	ListFocusActions ListFocus = "actions"
)

type QueryChangeMsg string

func NewList(items ...types.ListItem) *List {
	filter := NewFilter()
	filter.DrawLines = true

	statusBar := NewStatusBar()

	input := textinput.New()
	input.Prompt = ""
	input.PlaceholderStyle = lipgloss.NewStyle().Faint(true)
	input.Placeholder = "Search Items..."

	viewport := viewport.Model{}
	viewport.Style = lipgloss.NewStyle().Padding(0, 1)

	list := &List{
		spinner:   spinner.New(),
		input:     input,
		filter:    filter,
		viewport:  viewport,
		statusBar: statusBar,
		focus:     ListFocusItems,
	}

	list.SetItems(items...)
	return list
}

func (l *List) ResetSelection() {
	l.filter.ResetSelection()

	if selection := l.filter.Selection(); selection != nil {
		l.statusBar.SetActions(selection.(ListItem).Actions...)
	} else {
		l.statusBar.SetActions(l.Actions...)
	}
}

func (c *List) updateViewport(detail types.ListItemDetail) {
	var content string

	if detail.Markdown != "" {
		if len(detail.Markdown) > 5_000 {
			detail.Markdown = detail.Markdown[:min(5_000, len(detail.Markdown))] + "\n\n**Content truncated**"
		}
		style := AnsiStyle()
		style.Document.Margin = nil
		render, err := glamour.NewTermRenderer(
			glamour.WithStyles(style),
			glamour.WithWordWrap(c.viewport.Width-2),
		)
		if err != nil {
			c.viewport.SetContent(err.Error())
			return
		}

		content, err = render.Render(detail.Markdown)
		if err != nil {
			c.viewport.SetContent(err.Error())
			return
		}
	} else {
		if len(detail.Text) > 5_000 {
			detail.Text = detail.Text[:min(5_000, len(detail.Text))] + "\n\n**Content truncated**"
		}
		content = wrap.String(wordwrap.String(utils.StripAnsi(detail.Text), c.viewport.Width-2), c.viewport.Width-2)
		content = lipgloss.NewStyle().Padding(0, 2).Render(content)
	}

	c.viewport.GotoTop()
	c.viewport.SetContent(content)
}

func (l *List) SetActions(actions ...types.Action) {
	l.Actions = actions
	if l.filter.Selection() == nil {
		l.statusBar.SetActions(actions...)
	}
}

func (l *List) SetEmptyText(text string) {
	l.filter.EmptyText = text
}

func (c *List) Init() tea.Cmd {
	if c.focus == ListFocusActions {
		c.focus = ListFocusItems
		c.input.Placeholder = "Search Items..."
		c.input.SetValue(c.query)
	}
	return c.input.Focus()
}

func (c *List) Focus() tea.Cmd {
	c.statusBar.Reset()
	c.focus = ListFocusItems
	c.input.Placeholder = "Search Items..."
	c.input.SetValue(c.query)

	return c.input.Focus()
}

func (c *List) Blur() tea.Cmd {
	return nil
}

func (c *List) SetQuery(query string) tea.Cmd {
	c.input.SetValue(query)

	if c.focus == ListFocusItems {
		c.query = query
		if c.OnQueryChange != nil {
			return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				c.filter.EmptyText = "Loading..."
				if query == c.input.Value() {
					return QueryChangeMsg(query)
				}

				return nil
			})
		}

		c.FilterItems(query)
		c.filter.ResetSelection()
	} else {
		c.statusBar.FilterActions(query)
	}

	return nil
}

func (c *List) FilterItems(query string) {
	c.filter.FilterItems(query)
	selection := c.filter.Selection()
	if selection == nil {
		c.statusBar.SetActions(c.Actions...)
	} else {
		listItem := selection.(ListItem)
		c.statusBar.SetActions(listItem.Actions...)

		if c.showDetail {
			c.updateViewport(listItem.Detail)
		}
	}
}

func (c *List) SetShowDetail(showDetail bool) {
	c.showDetail = showDetail
	c.SetSize(c.width, c.height)

	if selection := c.filter.Selection(); selection != nil {
		c.updateViewport(selection.(ListItem).Detail)
	}
}

func (c *List) SetSize(width, height int) {
	c.width, c.height = width, height
	availableHeight := max(0, height-4)

	c.statusBar.Width = width

	if c.showDetail {
		third := width / 3
		c.filter.SetSize(third, availableHeight)
		c.viewport.Width = third * 2
		if width%3 == 0 {
			c.viewport.Width -= 1
		}

		c.viewport.Height = availableHeight
	} else {
		c.filter.SetSize(width, availableHeight)
	}
}

func (c List) Selection() (types.ListItem, bool) {
	selection := c.filter.Selection()
	if selection == nil {
		return types.ListItem{}, false
	}

	item := selection.(ListItem)
	return types.ListItem(item), true
}

func (c *List) SetItems(items ...types.ListItem) {
	filterItems := make([]FilterItem, len(items))
	for i, item := range items {
		filterItems[i] = ListItem(item)
	}

	c.filter.SetItems(filterItems...)

	if c.OnQueryChange == nil {
		c.FilterItems(c.Query())
	}
}

func (c *List) SetIsLoading(isLoading bool) tea.Cmd {
	c.isLoading = isLoading
	if isLoading {
		return c.spinner.Tick
	}
	return nil
}

func (c List) Query() string {
	return c.query
}

func (c *List) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if c.statusBar.expanded {
				c.focus = ListFocusItems
				c.input.SetValue(c.query)
				c.input.Placeholder = "Search Items..."

				c.statusBar.Reset()
				return c, nil
			}

			if c.input.Value() != "" {
				return c, c.SetQuery("")
			}

			return c, PopPageCmd
		case "ctrl+j":
			if !c.showDetail {
				break
			}

			c.viewport.LineDown(1)
			return c, nil
		case "ctrl+k":
			if !c.showDetail {
				break
			}

			c.viewport.LineUp(1)
			return c, nil
		case "tab":
			if c.statusBar.expanded {
				break
			}

			c.input.SetValue("")
			c.input.Placeholder = "Search Actions..."
			c.statusBar.expanded = true
			c.focus = ListFocusActions
			return c, nil
		case "right", "left":
			if c.statusBar.expanded {
				statusBar, cmd := c.statusBar.Update(msg)
				c.statusBar = statusBar
				return c, cmd
			}

			input, cmd := c.input.Update(msg)
			c.input = input
			return c, cmd
		}
	case QueryChangeMsg:
		if c.OnQueryChange == nil {
			return c, nil
		}

		if string(msg) != c.input.Value() {
			return c, nil
		}

		return c, c.OnQueryChange(string(msg))
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	statusBar, cmd := c.statusBar.Update(msg)
	c.statusBar = statusBar
	if cmd != nil {
		return c, cmd
	}

	input, cmd := c.input.Update(msg)
	if input.Value() != c.input.Value() {
		cmds = append(cmds, c.SetQuery(input.Value()))
	}

	c.input = input
	cmds = append(cmds, cmd)

	filter, cmd := c.filter.Update(msg)
	oldSelection := c.filter.Selection()
	newSelection := filter.Selection()
	if newSelection == nil {
		c.statusBar.SetActions(c.Actions...)
		c.updateViewport(types.ListItemDetail{})
	} else if oldSelection == nil || oldSelection.ID() != newSelection.ID() {
		listItem := newSelection.(ListItem)

		if c.showDetail {
			c.updateViewport(listItem.Detail)
		}

		c.statusBar.SetActions(newSelection.(ListItem).Actions...)
	}
	c.filter = filter
	cmds = append(cmds, cmd)

	if c.isLoading {
		c.spinner, cmd = c.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

func (c List) View() string {
	var headerRow string
	if c.isLoading {
		headerRow = fmt.Sprintf(" %s %s", c.spinner.View(), c.input.View())
	} else {
		headerRow = fmt.Sprintf("   %s", c.input.View())
	}

	var mainView string
	if c.showDetail {
		var bars []string
		for i := 0; i < c.height-4; i++ {
			bars = append(bars, "│")
		}
		vertical := strings.Join(bars, "\n")

		mainView = lipgloss.JoinHorizontal(lipgloss.Top, c.filter.View(), vertical, c.viewport.View())
	} else {
		mainView = c.filter.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, headerRow, separator(c.width), mainView, c.statusBar.View())
}
