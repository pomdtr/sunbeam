package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
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
		title = styles.Title.Copy().Foreground(accentColor).Render(fmt.Sprintf("> %s", i.Title))
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

		return fmt.Sprintf("%s%s", title, accessories)
	} else if lipgloss.Width(title+subtitle+accessories) > width {
		availableWidth := width - lipgloss.Width(title+accessories)
		subtitle := subtitle[:utils.Max(0, availableWidth)]
		return fmt.Sprintf("%s%s%s", title, subtitle, accessories)
	} else {
		blankWidth := width - lipgloss.Width(title) - lipgloss.Width(subtitle) - lipgloss.Width(accessories)
		blanks := strings.Repeat(blank, blankWidth)
		return fmt.Sprintf("%s%s%s%s", title, subtitle, blanks, accessories)
	}

}

type List struct {
	width, height int

	footer    *Footer
	spinner   spinner.Model
	actions   *ActionList
	isLoading bool

	filter *Filter
}

func NewList(title string) *List {
	viewport := viewport.New(0, 0)
	viewport.MouseWheelEnabled = true

	actions := NewActionList()

	spinner := spinner.NewModel()
	spinner.Style = styles.Background.Copy().Foreground(accentColor).Padding(0, 1)

	filter := NewFilter()
	filter.DrawLines = true
	filter.TextInput.Focus()

	footer := NewFooter(title)

	return &List{
		isLoading: true,
		actions:   actions,
		spinner:   spinner,
		filter:    filter,
		footer:    footer,
	}
}

func (c *List) Init() tea.Cmd {
	return tea.Batch(c.spinner.Tick)
}

func (c *List) SetSize(width, height int) {
	availableHeight := height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footer.View())
	c.width, c.height = width, availableHeight
	c.footer.Width = width
	c.filter.SetSize(width, availableHeight)
	c.actions.SetSize(width, height)
}

func (c *List) SetItems(items []ListItem) {
	c.isLoading = false

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

func (c List) headerView() string {
	var headerRow string
	if c.isLoading {
		spinner := c.spinner.View()
		textInput := styles.Text.Copy().Width(c.width - lipgloss.Width(spinner)).Render("Loading...")
		headerRow = lipgloss.JoinHorizontal(lipgloss.Top, c.spinner.View(), textInput)
	} else {
		headerRow = styles.Text.Copy().PaddingLeft(3).Width(c.width).Render(c.filter.TextInput.View())
	}

	line := strings.Repeat("─", c.width)
	line = styles.Title.Render(line)
	return lipgloss.JoinVertical(lipgloss.Left, headerRow, line)
}

func (c *List) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case FilterItemChange:
		if c.actions.Shown {
			return c, nil
		}
		c.updateActions()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			if c.actions.Shown {
				c.actions.Hide()
			} else if c.filter.TextInput.Value() != "" {
				c.filter.TextInput.SetValue("")
				c.filter.FilterItems("")
			} else {
				return c, PopCmd
			}
		}
	case ReloadMsg:
		c.isLoading = true
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	c.spinner, cmd = c.spinner.Update(msg)
	cmds = append(cmds, cmd)

	c.actions, cmd = c.actions.Update(msg)
	cmds = append(cmds, cmd)
	if !c.actions.Shown {
		c.filter, cmd = c.filter.Update(msg)
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

type ListDetailOutputMsg string

func (c List) View() string {
	if c.actions.Shown {
		return c.actions.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.filter.View(), c.footer.View())
}

func (c List) Query() string {
	return c.filter.TextInput.Value()
}

func NewErrorCmd(err error) func() tea.Msg {
	return func() tea.Msg {
		return err
	}
}
