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
		actions[i] = NewAction(extensionName, scriptAction)
	}

	return ListItem{
		Title:       item.Title,
		Subtitle:    item.Subtitle,
		Detail:      item.Detail.Command,
		Actions:     actions,
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

	prompt := "  "
	if selected {
		prompt = "> "
	}
	accessories := strings.Join(i.Accessories, " • ")
	blank := " "
	title := DefaultStyles.Primary.Render(i.Title)
	subtitle := DefaultStyles.Secondary.Render(i.Subtitle)

	var itemView string
	// No place to display the accessories, just return the title
	if lipgloss.Width(accessories) > width {
		view := fmt.Sprintf("%s%s", prompt, title)
		itemView = view[:utils.Min(lipgloss.Width(view), width)]
	} else if lipgloss.Width(prompt+title+blank+accessories) > width {
		availableWidth := width - lipgloss.Width(prompt+blank+accessories)
		title := title[:utils.Max(0, availableWidth)]
		itemView = fmt.Sprintf("%s%s%s%s", prompt, title, blank, accessories)
	} else if lipgloss.Width(prompt+title+blank+subtitle+blank+accessories) > width {
		availableWidth := width - lipgloss.Width(prompt+blank+title+blank+accessories)
		subtitle := subtitle[:utils.Max(0, availableWidth)]
		itemView = fmt.Sprintf("%s%s%s%s%s%s", prompt, title, blank, subtitle, blank, accessories)
	} else {
		blankWidth := width - lipgloss.Width(prompt+title+blank+subtitle+accessories)
		blanks := strings.Repeat(" ", blankWidth)
		itemView = fmt.Sprintf("%s%s%s%s%s%s", prompt, title, blank, subtitle, blanks, accessories)
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
	t.Prompt = "  "
	t.Placeholder = "Search..."
	t.PlaceholderStyle = DefaultStyles.Secondary
	t.TextStyle = DefaultStyles.Primary

	footer := NewFooter()
	viewport := viewport.New(0, 0)
	viewport.MouseWheelEnabled = true

	filter := Filter{
		viewport:   &viewport,
		itemHeight: 2,
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
	c.filter.Filter(c.textInput.Value())
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
	selectedItem := c.filter.Selection()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			if c.textInput.Value() != "" {
				c.textInput.SetValue("")
				c.filter.Filter("")
			} else {
				return c, PopCmd
			}
		case tea.KeyEnter:
			if selectedItem == nil {
				break
			}
			selectedItem, _ := selectedItem.(ListItem)
			if len(selectedItem.Actions) == 0 {
				break
			}
			return c, selectedItem.Actions[0].Exec()
		default:
			if selectedItem == nil {
				break
			}
			selectedItem, _ := selectedItem.(ListItem)
			for _, action := range selectedItem.Actions {
				if action.Shortcut == msg.String() {
					return c, action.Exec()
				}
			}
		}
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd
	t, cmd := c.textInput.Update(msg)
	cmds = append(cmds, cmd)
	if t.Value() != c.textInput.Value() {
		c.filter.Filter(t.Value())
	}
	c.textInput = &t

	c.filter, cmd = c.filter.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

type ListDetailOutputMsg string

func (c *List) View() string {
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

type ReloadMsg struct {
	input api.ScriptInput
}

func NewReloadCmd(input api.ScriptInput) func() tea.Msg {
	return func() tea.Msg {
		return ReloadMsg{
			input: input,
		}
	}
}

func NewErrorCmd(err error) func() tea.Msg {
	return func() tea.Msg {
		return err
	}
}
