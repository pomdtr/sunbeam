package tui

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/sahilm/fuzzy"
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

func (i ListItem) String() string {
	if i.Subtitle == "" {
		return i.Title
	}
	return fmt.Sprintf("%s %s", i.Title, i.Subtitle)
}

func (i ListItem) View(width int) string {
	if width == 0 {
		return ""
	}
	accessories := strings.Join(i.Accessories, " • ")
	titleWidth := lipgloss.Width(i.Title)
	subtitleWidth := lipgloss.Width(i.Subtitle)
	accessoriesWidth := lipgloss.Width(accessories)

	// No place to display the accessories, just return the title
	if accessoriesWidth > width {
		title := i.Title[:utils.Min(titleWidth, width)]
		return DefaultStyles.Primary.Render(title)
	}

	if titleWidth+1+accessoriesWidth > width {
		availableWidth := width - accessoriesWidth - 1
		title := i.Title[:utils.Min(titleWidth, availableWidth)]

		return lipgloss.JoinHorizontal(lipgloss.Top, DefaultStyles.Primary.Render(title), " ", DefaultStyles.Secondary.Render(accessories))
	}

	if titleWidth+1+subtitleWidth+1+accessoriesWidth > width {
		availableWidth := width - titleWidth - accessoriesWidth - 2
		subtitle := i.Subtitle[:availableWidth]

		return lipgloss.JoinHorizontal(lipgloss.Top, DefaultStyles.Primary.Render(i.Title), " ", DefaultStyles.Secondary.Render(subtitle), " ", DefaultStyles.Secondary.Render(accessories))
	}

	blankWidth := width - titleWidth - 1 - subtitleWidth - accessoriesWidth
	blanks := strings.Repeat(" ", blankWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, DefaultStyles.Primary.Render(i.Title), " ", DefaultStyles.Secondary.Render(i.Subtitle), blanks, DefaultStyles.Secondary.Render(accessories))
}

type List struct {
	width, height int

	textInput *textinput.Model
	footer    *Footer
	viewport  *viewport.Model

	dynamic bool

	startIndex, selectedIndex int
	items                     []ListItem
	filteredItems             []ListItem
}

func NewList(dynamic bool) *List {
	t := textinput.New()
	t.Prompt = "  "
	t.Placeholder = "Search..."

	v := viewport.New(0, 0)
	f := NewFooter()

	return &List{
		textInput: &t,
		viewport:  &v,
		dynamic:   dynamic,
		footer:    f,
	}
}

func (c *List) Init() tea.Cmd {
	return tea.Batch(c.textInput.Focus(), c.updateIndexes(0))
}

func (c *List) SetItems(items []ListItem) {
	c.items = items
	c.filteredItems = filterItems(c.textInput.Value(), items)
	c.updateIndexes(0)
}

func (c *List) SetSize(width, height int) {
	availableHeight := height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footer.View()) - 1

	splitSize := width / 2
	c.viewport.Width = splitSize
	c.viewport.Height = availableHeight

	c.width, c.height = width, availableHeight
	c.footer.Width = width
	c.updateIndexes(c.selectedIndex)
}

func (c List) SelectedItem() *ListItem {
	if c.selectedIndex >= len(c.filteredItems) {
		return nil
	}
	return &c.filteredItems[c.selectedIndex]
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

func (c *List) filterItems(term string) tea.Cmd {
	c.filteredItems = filterItems(term, c.items)
	return c.updateIndexes(0)
}

func (c *List) Update(msg tea.Msg) (*List, tea.Cmd) {
	var cmds []tea.Cmd

	selectedItem := c.SelectedItem()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyDown, tea.KeyTab, tea.KeyCtrlJ:
			if c.selectedIndex < len(c.filteredItems)-1 {
				cmd := c.updateIndexes(c.selectedIndex + 1)

				return c, cmd
			}
		case tea.KeyUp, tea.KeyShiftTab, tea.KeyCtrlK:
			if c.selectedIndex > 0 {
				cmd := c.updateIndexes(c.selectedIndex - 1)
				return c, cmd
			}
		case tea.KeyEscape:
			if c.textInput.Value() != "" {
				c.textInput.SetValue("")
				cmd := c.filterItems("")
				return c, cmd
			} else {
				return c, PopCmd
			}
		case tea.KeyEnter:
			if selectedItem == nil || len(selectedItem.Actions) == 0 {
				break
			}
			return c, selectedItem.Actions[0].Exec()
		default:
			if selectedItem == nil {
				break
			}
			for _, action := range selectedItem.Actions {
				if action.Shortcut == msg.String() {
					return c, action.Exec()
				}
			}
		}
	case ListDetailOutputMsg:
		c.viewport.SetContent(string(msg))
		return c, nil

	case DebounceMsg:
		if msg.check() {
			return c, msg.cmd
		}
	}

	t, cmd := c.textInput.Update(msg)
	cmds = append(cmds, cmd)
	if t.Value() != c.textInput.Value() {
		if c.dynamic {
			cmds = append(cmds, tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg {
				query := t.Value()
				return DebounceMsg{
					check: func() bool {
						return query == c.Query()
					},
					cmd: NewQueryUpdateCmd(t.Value()),
				}
			}))
		} else {
			cmd := c.filterItems(t.Value())
			cmds = append(cmds, cmd)
		}
	}
	c.textInput = &t

	return c, tea.Batch(cmds...)
}

type ListDetailOutputMsg string

func (c *List) updateIndexes(selectedIndex int) tea.Cmd {
	for selectedIndex < c.startIndex {
		c.startIndex--
	}
	for selectedIndex > c.startIndex+c.height {
		c.startIndex++
	}
	c.selectedIndex = selectedIndex
	selectedItem := c.SelectedItem()

	if selectedItem == nil {
		return nil
	}

	c.footer.SetActions(selectedItem.Actions)

	if selectedItem.Detail == "" {
		return nil
	}

	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg {
		return DebounceMsg{
			check: func() bool {
				return c.SelectedItem() == selectedItem
			},
			cmd: func() tea.Msg {
				res, err := exec.Command("sh", "-c", selectedItem.Detail).Output()
				if err != nil {
					return NewErrorCmd("Error running command: %s", err)
				}
				return ListDetailOutputMsg(res)
			},
		}
	})
}

func (c *List) listView(width int) string {
	rows := make([]string, 0)
	items := c.filteredItems

	endIndex := utils.Min(c.startIndex+c.height+1, len(items))
	for i := c.startIndex; i < endIndex; i++ {
		var prompt string
		if i == c.selectedIndex {
			prompt = "> "
		} else {
			prompt = "  "
		}
		itemWidth := utils.Max(width-1, 0)
		itemView := lipgloss.JoinHorizontal(lipgloss.Top, prompt, items[i].View(itemWidth))
		rows = append(rows, itemView)
	}

	for i := len(rows) - 1; i < c.height; i++ {
		rows = append(rows, strings.Repeat(" ", width))
	}
	return strings.Join(rows, "\n")
}

func (c *List) View() string {
	var embed string
	if c.SelectedItem() != nil && c.SelectedItem().Detail != "" {
		availableWidth := c.width - c.viewport.Width - 3
		separator := make([]string, c.height+1)
		for i := 0; i <= c.height; i++ {
			separator[i] = " │ "
		}
		embed = lipgloss.JoinHorizontal(lipgloss.Top, c.listView(availableWidth), strings.Join(separator, "\n"), c.viewport.View())
	} else {
		embed = c.listView(c.width)
	}
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), embed, c.footer.View())
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

// filterItems uses the sahilm/fuzzy to filter through the list.
// This is set by default.
func filterItems(term string, items []ListItem) []ListItem {
	if term == "" {
		return items
	}
	targets := make([]string, len(items))
	for i, item := range items {
		targets[i] = item.String()
	}
	var ranks = fuzzy.Find(term, targets)
	sort.Stable(ranks)
	filteredItems := make([]ListItem, len(ranks))
	for i, r := range ranks {
		item := items[r.Index]
		filteredItems[i] = item
	}
	return filteredItems
}

func NewErrorCmd(format string, values ...any) func() tea.Msg {
	return func() tea.Msg {
		return fmt.Errorf(format, values...)
	}
}
