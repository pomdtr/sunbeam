package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/utils"
)

type ListItem struct {
	Title       string
	Subtitle    string
	Accessories []string
	Actions     []Action
}

func (i ListItem) FilterValue() string {
	if i.Subtitle == "" {
		return i.Title
	}
	return fmt.Sprintf("%s %s", i.Title, i.Subtitle)
}

func (i ListItem) Render(width int, selected bool) string {
	if width == 0 {
		return ""
	}

	var title string
	if selected {
		title = styles.Title.Copy().Foreground(accentColor).PaddingRight(1).Render(fmt.Sprintf("> %s", i.Title))
	} else {
		title = styles.Title.Copy().PaddingRight(1).Render(fmt.Sprintf("  %s", i.Title))
	}

	blank := styles.Background.Render(" ")
	accessories := strings.Join(i.Accessories, " • ")
	accessories = styles.Text.Render(accessories)
	subtitle := styles.Text.Render(i.Subtitle)

	if lipgloss.Width(accessories) > width {
		return title
	} else if lipgloss.Width(title)+lipgloss.Width(accessories) > width {
		availableWidth := width - lipgloss.Width(accessories)

		title := title[:utils.Max(0, availableWidth)]

		return lipgloss.JoinHorizontal(lipgloss.Top, title, accessories)
	} else if lipgloss.Width(title+subtitle+accessories) > width {
		availableWidth := width - lipgloss.Width(title+accessories)
		subtitle := subtitle[:utils.Max(0, availableWidth)]
		return lipgloss.JoinHorizontal(lipgloss.Top, title, subtitle, accessories)
	} else {
		blankWidth := width - lipgloss.Width(title) - lipgloss.Width(subtitle) - lipgloss.Width(accessories)
		blanks := strings.Repeat(blank, blankWidth)
		return lipgloss.JoinHorizontal(lipgloss.Top, title, subtitle, blanks, accessories)
	}
}

type List struct {
	header  Header
	footer  Footer
	actions *ActionList
	dynamic bool

	filter *Filter
}

func NewList(title string) *List {
	actions := NewActionList()

	header := NewHeader()
	header.SetIsLoading(true)

	filter := NewFilter()
	filter.DrawLines = true

	footer := NewFooter(title)

	return &List{
		actions: actions,
		header:  header,
		filter:  filter,
		footer:  footer,
	}
}

func (c *List) Init() tea.Cmd {
	return tea.Batch(c.header.Init(), c.header.input.Focus())

}

func (c *List) SetSize(width, height int) {
	availableHeight := height - lipgloss.Height(c.header.View()) - lipgloss.Height(c.footer.View())
	c.footer.Width = width
	c.header.Width = width
	c.filter.SetSize(width, availableHeight)
	c.actions.SetSize(width, height)
}

func (c *List) SetItems(items []ListItem) {
	c.header.SetIsLoading(false)
	filterItems := make([]FilterItem, len(items))
	for i, item := range items {
		filterItems[i] = item
	}

	c.filter.SetItems(filterItems)
	c.updateActions()
}

func (l *List) updateActions() {
	if l.filter.Selection() == nil {
		l.actions.SetTitle("")
		l.actions.SetActions()
		l.footer.SetBindings()
	}
	item, _ := l.filter.Selection().(ListItem)
	l.actions.SetTitle(item.Title)
	l.actions.SetActions(item.Actions...)
	if len(item.Actions) > 0 {
		l.footer.SetBindings(
			key.NewBinding(key.WithKeys(item.Actions[0].Shortcut), key.WithHelp("↩", item.Actions[0].Title)),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Show Actions")),
		)
	} else {
		l.footer.SetBindings()
	}
}

func (c *List) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case FilterItemChange:
		if c.actions.Focused() {
			return c, nil
		}
		c.updateActions()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			if c.actions.Focused() {
				break
			} else if c.header.input.Value() != "" {
				c.header.input.SetValue("")
				c.filter.FilterItems("")
			} else {
				return c, PopCmd
			}
		}
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	c.actions, cmd = c.actions.Update(msg)
	cmds = append(cmds, cmd)

	if c.actions.Focused() {
		return c, tea.Batch(cmds...)
	}

	header, cmd := c.header.Update(msg)
	cmds = append(cmds, cmd)
	if header.Value() != c.header.Value() {
		if c.dynamic {
			cmd = func() tea.Msg {
				return ReloadMsg{
					Params: map[string]any{
						"query": header.Value(),
					},
				}
			}
		} else {
			cmd = c.filter.FilterItems(header.Value())
		}
		cmds = append(cmds, cmd)
	}
	c.header = header

	c.filter, cmd = c.filter.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

type ListDetailOutputMsg string

func (c List) View() string {
	if c.actions.Focused() {
		return c.actions.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, c.header.View(), c.filter.View(), c.footer.View())
}

func (c List) Query() string {
	return c.header.input.Value()
}

func NewErrorCmd(err error) func() tea.Msg {
	return func() tea.Msg {
		return err
	}
}
