package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	var titleStyle lipgloss.Style
	if selected {
		title = fmt.Sprintf("> %s", i.Title)
		titleStyle = styles.Title.Copy().Foreground(accentColor)
	} else {
		title = fmt.Sprintf("  %s", i.Title)
		titleStyle = styles.Title.Copy()
	}

	subtitle := fmt.Sprintf(" %s", i.Subtitle)
	var blanks string
	accessories := fmt.Sprintf(" %s", strings.Join(i.Accessories, " • "))

	if width >= len(title+subtitle+accessories) {
		availableWidth := width - len(title+subtitle+accessories)
		blanks = strings.Repeat(" ", availableWidth)
	} else if width >= len(title+accessories) {
		subtitle = subtitle[:width-len(title+accessories)]
	} else if width >= len(accessories) {
		subtitle = ""
		title = title[:width-len(" "+accessories)]
	} else {
		accessories = ""
		title = title[:width]
	}

	title = titleStyle.Render(title)
	subtitle = styles.Text.Copy().Render(subtitle)
	blanks = styles.Text.Render(blanks)
	accessories = styles.Text.Copy().Italic(true).Render(accessories)

	return lipgloss.JoinHorizontal(lipgloss.Top, title, subtitle, blanks, accessories)
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
