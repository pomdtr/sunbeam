package pages

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/bubbles"
	"github.com/pomdtr/sunbeam/bubbles/list"
	"github.com/pomdtr/sunbeam/scripts"
	"github.com/pomdtr/sunbeam/utils"
)

type ListContainer struct {
	list      *list.Model
	textInput *textinput.Model
	width     int
	height    int
	response  *scripts.ListResponse
}

func NewListContainer(res *scripts.ListResponse) Container {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	listItems := make([]list.Item, len(res.Items))
	for i, item := range res.Items {
		listItems[i] = item
	}

	t := textinput.New()
	t.Prompt = ""
	t.Placeholder = res.SearchBarPlaceholder
	if res.SearchBarPlaceholder != "" {
		t.Placeholder = res.SearchBarPlaceholder
	} else {
		t.Placeholder = "Search..."
	}

	return &ListContainer{
		list:      &l,
		textInput: &t,
		response:  res,
	}
}

func (c *ListContainer) SetSize(width, height int) {
	c.width, c.height = width, height
	c.list.SetSize(width, height-lipgloss.Height(c.footerView())-lipgloss.Height(c.headerView()))
}

func (c *ListContainer) headerView() string {
	input := c.textInput.View()
	line := strings.Repeat("─", c.width)
	return lipgloss.JoinVertical(lipgloss.Left, input, line)
}

func (c *ListContainer) footerView() string {
	selectedItem := c.list.SelectedItem()
	if selectedItem == nil {
		return bubbles.SunbeamFooter(c.width, c.response.Title)
	}

	if item, ok := selectedItem.(scripts.ScriptItem); ok && len(item.Actions) > 0 {
		return bubbles.SunbeamFooterWithActions(c.width, c.response.Title, item.Actions[0].Title())
	} else {
		return bubbles.SunbeamFooter(c.width, c.response.Title)
	}
}

func (c *ListContainer) Update(msg tea.Msg) (Container, tea.Cmd) {
	var cmds []tea.Cmd
	log.Printf("ListPage.Update: %T, %v", msg, msg)

	selectedItem := c.list.SelectedItem()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if selectedItem == nil {
				break
			}
			selectedItem := selectedItem.(scripts.ScriptItem)
			return c, utils.SendMsg(selectedItem.Actions[0])
		case tea.KeyEscape:
			return c, PopCmd
		default:
			if selectedItem == nil {
				break
			}
			selectedItem := selectedItem.(scripts.ScriptItem)
			for _, action := range selectedItem.Actions {
				if action.Keybind == msg.String() {
					return c, utils.SendMsg(action)
				}
			}
		}
	case scripts.ScriptResponse:
		items := make([]list.Item, len(msg.List.Items))
		for i, item := range msg.List.Items {
			items[i] = item
		}
		cmd := c.list.SetItems(items)
		return c, cmd
	}

	t, cmd := c.textInput.Update(msg)
	if c.response.OnQueryChange != nil && t.Value() != c.textInput.Value() {
		cmds = append(cmds, utils.SendMsg(*c.response.OnQueryChange))
	}
	cmds = append(cmds, cmd)
	c.textInput = &t

	l, cmd := c.list.Update(msg)
	cmds = append(cmds, cmd)
	c.list = &l

	return c, tea.Batch(cmds...)
}

func (c *ListContainer) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.list.View(), c.footerView())
}
