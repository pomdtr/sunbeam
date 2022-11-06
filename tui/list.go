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
		title = DefaultStyles.Selection.Render(fmt.Sprintf("> %s", i.Title))
	} else {
		title = DefaultStyles.Primary.Render(fmt.Sprintf("  %s", i.Title))
	}
	accessories := strings.Join(i.Accessories, " • ")
	accessories = DefaultStyles.Secondary.Render(accessories)
	blank := " "
	subtitle := DefaultStyles.Secondary.Render(i.Subtitle)

	if lipgloss.Width(accessories) > width {
		return title[:utils.Min(lipgloss.Width(title), width)]
	} else if lipgloss.Width(title+blank+accessories) > width {
		availableWidth := width - lipgloss.Width(blank+accessories)
		title := title[:utils.Max(0, availableWidth)]
		return fmt.Sprintf("%s%s%s", title, blank, accessories)
	} else if lipgloss.Width(title+blank+subtitle+blank+accessories) > width {
		availableWidth := width - lipgloss.Width(title+blank+accessories)
		subtitle := subtitle[:utils.Max(0, availableWidth)]
		return fmt.Sprintf("%s%s%s%s%s", title, blank, subtitle, blank, accessories)
	} else {
		blankWidth := width - lipgloss.Width(title+blank+subtitle+accessories)
		blanks := strings.Repeat(" ", blankWidth)
		return fmt.Sprintf("%s%s%s%s%s", title, blank, subtitle, blanks, accessories)
	}

}

type List struct {
	width, height int

	footer    Footer
	spinner   spinner.Model
	actions   ActionList
	isLoading bool

	filter Filter
}

func NewList(title string) *List {
	viewport := viewport.New(0, 0)
	viewport.MouseWheelEnabled = true

	actions := NewActionList()

	spinner := spinner.NewModel()

	filter := NewFilter()
	filter.DrawLines = true

	footer := NewFooter(title)
	footer.SetBindings(
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", "Select")),
		key.NewBinding(key.WithKeys("ctrl+h"), key.WithHelp("⇥", "Actions")),
	)

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

	if c.filter.Selection() != nil {
		item, _ := c.filter.Selection().(ListItem)
		c.actions.SetActions(item.Title, item.Actions...)
	}
}

func (c List) headerView() string {
	var headerRow string
	if c.isLoading {
		headerRow = lipgloss.JoinHorizontal(lipgloss.Top, " ", c.spinner.View(), " ", c.filter.Model.View())
	} else {
		headerRow = lipgloss.JoinHorizontal(lipgloss.Top, "   ", c.filter.Model.View())
	}

	line := strings.Repeat("─", c.width)
	return lipgloss.JoinVertical(lipgloss.Left, headerRow, line)
}

type DebounceMsg struct {
	check func() bool
	cmd   tea.Cmd
}

func (c *List) Update(msg tea.Msg) (Container, tea.Cmd) {
	switch msg := msg.(type) {
	case FilterItemChange:
		if c.actions.Shown {
			return c, nil
		}
		if msg.FilterItem == nil {
			c.actions.SetActions("")
			return c, nil
		}
		listItem, _ := msg.FilterItem.(ListItem)
		c.actions.SetActions(listItem.Title, listItem.Actions...)
		return c, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			if c.actions.Shown {
				c.actions.Hide()
			} else if c.filter.Value() != "" {
				c.filter.SetValue("")
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
	return c.filter.Value()
}

func NewErrorCmd(err error) func() tea.Msg {
	return func() tea.Msg {
		return err
	}
}
