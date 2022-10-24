package tui

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/sahilm/fuzzy"
	"github.com/skratchdot/open-golang/open"
)

type ListContainer struct {
	width, height             int
	startIndex, selectedIndex int

	title         string
	textInput     *textinput.Model
	filteredItems []api.ListItem
	items         []api.ListItem
}

func NewListContainer(title string, items []api.ListItem, query string) Container {
	t := textinput.New()
	t.Prompt = "> "
	t.Placeholder = "Search..."
	t.SetValue(query)

	return &ListContainer{
		textInput:     &t,
		title:         title,
		filteredItems: filterItems(query, items),
		items:         items,
	}
}

// Rank defines a rank for a given item.
type Rank struct {
	// The index of the item in the original input.
	Index int
	// Indices of the actual word that were matched against the filter term.
	MatchedIndexes []int
}

// filterItems uses the sahilm/fuzzy to filter through the list.
// This is set by default.
func filterItems(term string, items []api.ListItem) []api.ListItem {
	if term == "" {
		return items
	}
	targets := make([]string, len(items))
	for i, item := range items {
		targets[i] = item.Title
	}
	var ranks = fuzzy.Find(term, targets)
	sort.Stable(ranks)
	filteredItems := make([]api.ListItem, len(ranks))
	for i, r := range ranks {
		filteredItems[i] = items[r.Index]
	}
	return filteredItems
}

func (c *ListContainer) Init() tea.Cmd {
	return c.textInput.Focus()
}

func (c *ListContainer) SetSize(width, height int) {
	c.width, c.height = width, height-lipgloss.Height(c.headerView())-lipgloss.Height(c.footerView())-1
	c.updateIndexes(c.selectedIndex)
}

func (c ListContainer) SelectedItem() *api.ListItem {
	if c.selectedIndex >= len(c.filteredItems) {
		return nil
	}
	return &c.filteredItems[c.selectedIndex]
}

func (c *ListContainer) headerView() string {
	input := c.textInput.View()
	line := strings.Repeat("─", c.width)
	return lipgloss.JoinVertical(lipgloss.Left, input, line)
}

func (c *ListContainer) footerView() string {
	selectedItem := c.SelectedItem()
	if selectedItem == nil {
		return SunbeamFooter(c.width, c.title)
	}

	if len(selectedItem.Actions) > 0 {
		return SunbeamFooterWithActions(c.width, c.title, selectedItem.Actions[0].Title)
	} else {
		return SunbeamFooter(c.width, c.title)
	}
}

func (c *ListContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	var cmds []tea.Cmd

	selectedItem := c.SelectedItem()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if selectedItem == nil || len(selectedItem.Actions) == 0 {
				break
			}
			return c, c.RunAction(selectedItem.Actions[0])
		case tea.KeyDown, tea.KeyTab, tea.KeyCtrlJ:
			if c.selectedIndex < len(c.filteredItems)-1 {
				c.updateIndexes(c.selectedIndex + 1)
			}
		case tea.KeyUp, tea.KeyShiftTab, tea.KeyCtrlK:
			if c.selectedIndex > 0 {
				c.updateIndexes(c.selectedIndex - 1)
			}
		case tea.KeyEscape:
			if c.textInput.Value() != "" {
				c.textInput.SetValue("")
			} else {
				return c, PopCmd
			}
		default:
			if selectedItem == nil {
				break
			}
			for _, action := range selectedItem.Actions {
				if action.Keybind == msg.String() {
					return c, c.RunAction(action)
				}
			}
		}
	}

	t, cmd := c.textInput.Update(msg)
	cmds = append(cmds, cmd)
	if t.Value() != c.textInput.Value() {
		c.filteredItems = filterItems(t.Value(), c.items)
	}
	c.textInput = &t

	return c, tea.Batch(cmds...)
}

func (c *ListContainer) updateIndexes(selectedIndex int) {
	for selectedIndex < c.startIndex {
		c.startIndex--
	}
	for selectedIndex > c.startIndex+c.height {
		c.startIndex++
	}
	c.selectedIndex = selectedIndex
}

func itemView(item api.ListItem, selected bool) string {
	var view string

	if item.Subtitle != "" {
		view = fmt.Sprintf("%s - %s", item.Title, item.Subtitle)
	} else {
		view = item.Title
	}
	if selected {
		return fmt.Sprintf("> %s", view)
	} else {
		return fmt.Sprintf("  %s", view)
	}
}

func (c *ListContainer) View() string {
	rows := make([]string, 0)
	items := c.filteredItems

	endIndex := utils.Min(c.startIndex+c.height+1, len(items))
	for i := c.startIndex; i < endIndex; i++ {
		rows = append(rows, itemView(items[i], i == c.selectedIndex))
	}

	for i := len(rows) - 1; i < c.height; i++ {
		rows = append(rows, "")
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), strings.Join(rows, "\n"), c.footerView())
}

func NewErrorCmd(format string, values ...any) func() tea.Msg {
	return func() tea.Msg {
		return fmt.Errorf(format, values...)
	}
}

type ReloadMsg struct {
	input api.CommandInput
}

func NewReloadCmd(input api.CommandInput) func() tea.Msg {
	return func() tea.Msg {
		return ReloadMsg{
			input: input,
		}
	}
}

func (l *ListContainer) RunAction(action api.ScriptAction) tea.Cmd {
	switch action.Type {
	case "push":
		command, ok := api.GetCommand(action.Target)
		if !ok {
			return NewErrorCmd("unknown command %s", action.Target)
		}

		input := api.NewCommandInput(action.Params)

		return NewPushCmd(NewRunContainer(command, input))
	case "reload":
		input := api.NewCommandInput(action.Params)
		input.Query = l.textInput.Value()
		return NewReloadCmd(input)
	case "exec":
		var cmd *exec.Cmd
		log.Printf("executing command: %s", action.Command)
		if len(action.Command) == 1 {
			cmd = exec.Command(action.Command[0])
		} else {
			cmd = exec.Command(action.Command[0], action.Command[1:]...)
		}
		_, err := cmd.Output()
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return NewErrorCmd("Unable to run cmd: %s", exitError.Stderr)
		}
		return tea.Quit
	case "open":
		var err error
		target := action.Path
		if target == "" {
			target = action.Url
		}
		if action.Application != "" {
			err = open.RunWith(target, action.Application)
		} else {
			err = open.Run(target)
		}
		if err != nil {
			return NewErrorCmd("failed to open: %s", target)
		}
		return tea.Quit
	case "copy":
		err := clipboard.WriteAll(action.Content)
		if err != nil {
			return NewErrorCmd("failed to copy %s to clipboard", err)
		}
		return tea.Quit
	default:
		log.Printf("Unknown action type: %s", action.Type)
		return tea.Quit
	}
}
