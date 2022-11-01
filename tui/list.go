package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
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

	textInput *textinput.Model
	footer    Footer
	spinner   spinner.Model
	isLoading bool

	filter Filter
}

func NewList() *List {
	t := textinput.New()
	t.Placeholder = "Search..."
	t.PlaceholderStyle = DefaultStyles.Secondary
	t.Prompt = ""
	t.TextStyle = DefaultStyles.Primary

	footer := NewFooter()
	viewport := viewport.New(0, 0)
	viewport.MouseWheelEnabled = true

	spinner := spinner.NewModel()

	filter := Filter{
		viewport:  &viewport,
		drawLines: true,
	}

	return &List{
		textInput: &t,
		isLoading: true,
		spinner:   spinner,
		filter:    filter,
		footer:    footer,
	}
}

func (c *List) Init() tea.Cmd {
	return tea.Batch(c.textInput.Focus(), c.spinner.Tick)
}

func (c *List) SetSize(width, height int) {
	availableHeight := height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footer.View())
	c.width, c.height = width, availableHeight
	c.footer.Width = width
	c.filter.SetSize(width, availableHeight)
}

func (c *List) SetItems(items []ListItem) {
	c.isLoading = false

	filterItems := make([]FilterItem, len(items))
	for i, item := range items {
		filterItems[i] = item
	}

	c.filter.SetItems(filterItems)
	c.filter.FilterItems(c.textInput.Value())
}

func (c List) headerView() string {
	var headerRow string
	if c.isLoading {
		headerRow = lipgloss.JoinHorizontal(lipgloss.Top, " ", c.spinner.View(), " ", c.textInput.View())
	} else {
		headerRow = lipgloss.JoinHorizontal(lipgloss.Top, "   ", c.textInput.View())
	}

	line := strings.Repeat("─", c.width)
	return lipgloss.JoinVertical(lipgloss.Left, headerRow, line)
}

type DebounceMsg struct {
	check func() bool
	cmd   tea.Cmd
}

func (c List) Update(msg tea.Msg) (Container, tea.Cmd) {
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
				return &c, PopCmd
			}
		case tea.KeyEnter:
			if !hasItemSelected || len(selectedItem.Actions) == 0 {
				break
			}
			return &c, selectedItem.Actions[0].SendMsg
		}
	case ReloadMsg:
		c.isLoading = true
	case RunMsg:
		if msg.Extension == "" {
			msg.Extension = selectedItem.Extension
		}
		return &c, NewPushCmd(NewRunContainer(msg.Extension, msg.Target, msg.Params))
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	c.spinner, cmd = c.spinner.Update(msg)
	cmds = append(cmds, cmd)

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

	return &c, tea.Batch(cmds...)
}

type ListDetailOutputMsg string

func (c List) View() string {
	filterSelection := c.filter.Selection()
	selectedItem, ok := filterSelection.(ListItem)
	if ok {
		c.footer.SetActions(selectedItem.Actions...)
	} else {
		c.footer.SetActions()
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.filter.View(), c.footer.View())
}

func (c List) Query() string {
	return c.textInput.Value()
}

func NewErrorCmd(err error) func() tea.Msg {
	return func() tea.Msg {
		return err
	}
}
