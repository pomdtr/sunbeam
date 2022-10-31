package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/utils"
)

type ListItem struct {
	Actions     []Action
	Extension   string
	Title       string
	Subtitle    string
	Accessories []string
	Detail      string
}

func NewListItem(extensionName string, item api.ListItem) ListItem {
	actions := make([]Action, len(item.Actions))

	for i, scriptAction := range item.Actions {
		if scriptAction.Shortcut == "" {
			if i == 0 {
				scriptAction.Shortcut = "enter"
			} else if i < 10 {
				scriptAction.Shortcut = fmt.Sprintf("ctrl+%d", i)
			}
		}
		actions[i] = NewAction(scriptAction)
	}

	return ListItem{
		Title:       item.Title,
		Subtitle:    item.Subtitle,
		Detail:      item.Detail.Command,
		Actions:     actions,
		Extension:   extensionName,
		Accessories: item.Accessories,
	}
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

	var itemView string
	// No place to display the accessories, just return the title
	if lipgloss.Width(accessories) > width {
		itemView = title[:utils.Min(lipgloss.Width(title), width)]
	} else if lipgloss.Width(title+blank+accessories) > width {
		availableWidth := width - lipgloss.Width(blank+accessories)
		title := title[:utils.Max(0, availableWidth)]
		itemView = fmt.Sprintf("%s%s%s", title, blank, accessories)
	} else if lipgloss.Width(title+blank+subtitle+blank+accessories) > width {
		availableWidth := width - lipgloss.Width(title+blank+accessories)
		subtitle := subtitle[:utils.Max(0, availableWidth)]
		itemView = fmt.Sprintf("%s%s%s%s%s", title, blank, subtitle, blank, accessories)
	} else {
		blankWidth := width - lipgloss.Width(title+blank+subtitle+accessories)
		blanks := strings.Repeat(" ", blankWidth)
		itemView = fmt.Sprintf("%s%s%s%s%s", title, blank, subtitle, blanks, accessories)
	}

	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("#9a9a9c")).Render(strings.Repeat("─", width))
	return lipgloss.JoinVertical(lipgloss.Left, itemView, separator)
}

type List struct {
	width, height int

	textInput *textinput.Model
	footer    Footer

	dynamic bool
	filter  Filter
}

func NewList(dynamic bool) *List {
	t := textinput.New()
	t.Prompt = "   "
	t.Placeholder = "Search..."
	t.PlaceholderStyle = DefaultStyles.Secondary
	t.TextStyle = DefaultStyles.Primary

	footer := NewFooter()
	viewport := viewport.New(0, 0)
	viewport.MouseWheelEnabled = true

	filter := Filter{
		viewport:   &viewport,
		ItemHeight: 2,
	}

	return &List{
		textInput: &t,
		dynamic:   dynamic,
		filter:    filter,
		footer:    footer,
	}
}

func (c *List) Init() tea.Cmd {
	return tea.Batch(c.textInput.Focus())
}

func (c *List) SetSize(width, height int) {
	availableHeight := height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footer.View())
	c.width, c.height = width, availableHeight
	c.footer.Width = width
	c.filter.SetSize(width, availableHeight)
}

func (c *List) SetItems(items []ListItem) {
	filterItems := make([]FilterItem, len(items))
	for i, item := range items {
		filterItems[i] = item
	}
	c.filter.SetItems(filterItems)
	if c.dynamic {
		return
	}
	c.filter.FilterItems(c.textInput.Value())
}

func (c *List) headerView() string {
	input := c.textInput.View()
	line := strings.Repeat("─", c.width)
	return lipgloss.JoinVertical(lipgloss.Left, input, line)
}

type DebounceMsg struct {
	check func() bool
	cmd   tea.Cmd
}

func (c *List) Update(msg tea.Msg) (*List, tea.Cmd) {
	filterSelection := c.filter.Selection()
	selectedItem, hasItemSelected := filterSelection.(ListItem)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			if c.textInput.Value() != "" {
				c.textInput.SetValue("")
				c.filter.FilterItems("")
			} else {
				return c, PopCmd
			}
		case tea.KeyEnter:
			if !hasItemSelected || len(selectedItem.Actions) == 0 {
				break
			}
			return c, selectedItem.Actions[0].SendMsg
		}
	case RunMsg:
		if !hasItemSelected {
			break
		}
		script, ok := api.Sunbeam.GetScript(selectedItem.Extension, msg.Target)
		if !ok {
			return c, NewErrorCmd(fmt.Errorf("No script found for %s", msg.Target))
		}
		return c, NewPushCmd(NewRunContainer(script, msg.Params))
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd
	c.footer, cmd = c.footer.Update(msg)
	cmds = append(cmds, cmd)

	t, cmd := c.textInput.Update(msg)
	cmds = append(cmds, cmd)
	if t.Value() != c.textInput.Value() {
		c.filter.FilterItems(t.Value())
	}
	c.textInput = &t

	c.filter, cmd = c.filter.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

type ListDetailOutputMsg string

func (c *List) View() string {
	filterSelection := c.filter.Selection()
	selectedItem, ok := filterSelection.(ListItem)
	if ok {
		c.footer.SetActions(selectedItem.Actions...)
	} else {
		c.footer.SetActions()
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.filter.View(), c.footer.View())
}

func (c *List) Query() string {
	return c.textInput.Value()
}

type QueryUpdateMsg struct {
	query string
}

func NewQueryUpdateCmd(query string) func() tea.Msg {
	return func() tea.Msg {
		return QueryUpdateMsg{
			query: query,
		}
	}
}

func NewErrorCmd(err error) func() tea.Msg {
	return func() tea.Msg {
		return err
	}
}
