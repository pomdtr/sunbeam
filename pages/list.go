package pages

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/bubbles"
	"github.com/pomdtr/sunbeam/scripts"
	"github.com/pomdtr/sunbeam/utils"
)

type ListContainer struct {
	textInput   *textinput.Model
	selectedIdx int
	width       int
	height      int
	response    *scripts.ListResponse
}

func NewListContainer(res *scripts.ListResponse) Container {
	t := textinput.New()
	t.Focus()
	t.Prompt = ""
	t.Placeholder = res.SearchBarPlaceholder
	if res.SearchBarPlaceholder != "" {
		t.Placeholder = res.SearchBarPlaceholder
	} else {
		t.Placeholder = "Search..."
	}

	return &ListContainer{
		textInput: &t,
		response:  res,
	}
}

func (c *ListContainer) SetSize(width, height int) {
	c.width, c.height = width, height
}

func (c ListContainer) SelectedItem() (scripts.ScriptItem, bool) {
	if c.selectedIdx < 0 || c.selectedIdx >= len(c.response.Items) {
		return scripts.ScriptItem{}, false
	}
	return c.response.Items[c.selectedIdx], true
}

func (c *ListContainer) headerView() string {
	input := c.textInput.View()
	line := strings.Repeat("─", c.width)
	return lipgloss.JoinVertical(lipgloss.Left, input, line)
}

func (c *ListContainer) footerView() string {
	selectedItem, ok := c.SelectedItem()
	if !ok {
		return bubbles.SunbeamFooter(c.width, c.response.Title)
	}

	if len(selectedItem.Actions) > 0 {
		return bubbles.SunbeamFooterWithActions(c.width, c.response.Title, selectedItem.Actions[0].Title)
	} else {
		return bubbles.SunbeamFooter(c.width, c.response.Title)
	}
}

func (c *ListContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	var cmds []tea.Cmd

	selectedItem, _ := c.SelectedItem()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return c, utils.SendMsg(selectedItem.Actions[0])
		case tea.KeyDown, tea.KeyTab, tea.KeyCtrlJ:
			if c.selectedIdx < len(c.response.Items)-1 {
				c.selectedIdx++
			}
		case tea.KeyUp, tea.KeyShiftTab, tea.KeyCtrlK:
			if c.selectedIdx > 0 {
				c.selectedIdx--
			}
		case tea.KeyEscape:
			return c, PopCmd
		default:
			for _, action := range selectedItem.Actions {
				if action.Keybind == msg.String() {
					return c, utils.SendMsg(action)
				}
			}
		}
	case scripts.ListResponse:
		c.response = &msg
		c.selectedIdx = 0
		return c, nil
	}

	t, cmd := c.textInput.Update(msg)
	cmds = append(cmds, cmd)
	c.textInput = &t

	return c, tea.Batch(cmds...)
}

func (c *ListContainer) View() string {
	rows := make([]string, 0)
	items := c.response.Items

	availableHeight := c.height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footerView())
	startIndex := utils.Max(0, c.selectedIdx-availableHeight+1)
	maxIndex := utils.Min(len(items), startIndex+availableHeight)

	for i := startIndex; i < maxIndex; i++ {
		if i == c.selectedIdx {
			rows = append(rows, fmt.Sprintf("> %s %s", items[i].Title, items[i].Subtitle))
		} else {
			rows = append(rows, fmt.Sprintf("  %s %s", items[i].Title, items[i].Subtitle))
		}
	}
	for i := 0; i < availableHeight-len(items); i++ {
		rows = append(rows, "")
	}

	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), strings.Join(rows, "\n"), c.footerView())
}
