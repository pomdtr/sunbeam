package pages

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jinzhu/copier"
	"github.com/pomdtr/sunbeam/bubbles"
	"github.com/pomdtr/sunbeam/bubbles/list"
	commands "github.com/pomdtr/sunbeam/commands"
	"github.com/pomdtr/sunbeam/utils"
)

type ActionList struct {
	list      *list.Model
	textInput *textinput.Model
	hidden    bool
	width     int
	height    int
}

func NewActionList(title string, actions []commands.ScriptAction) *ActionList {
	var l list.Model
	_ = copier.Copy(&l, &listContainer)

	textInput := textinput.NewModel()
	textInput.Prompt = ""
	textInput.Placeholder = "Search for actions..."
	textInput.Focus()

	actionList := ActionList{
		list:      &l,
		hidden:    true,
		textInput: &textInput,
	}
	actionList.setActions(actions)

	return &actionList
}

func (c *ActionList) setActions(actions []commands.ScriptAction) {
	listItems := make([]list.Item, len(actions))
	for i, action := range actions {
		listItems[i] = action
	}
	c.list.SetItems(listItems)
}

func (c *ActionList) Visible() bool {
	return !c.hidden
}

func (c *ActionList) Hide() {
	c.hidden = true
}

func (c *ActionList) Show() {
	c.hidden = false
}

func (c *ActionList) headerView() string {
	return bubbles.SunBeamHeaderWithInput(c.width, c.textInput)
}

func (c *ActionList) footerView() string {
	return bubbles.SunbeamFooter(c.width, "Actions")
}

func (c *ActionList) SetSize(width, height int) {
	c.width, c.height = width, height
	headerHeight := lipgloss.Height(c.headerView())
	footerHeight := lipgloss.Height(c.footerView())
	c.list.SetSize(c.width, c.height-headerHeight-footerHeight)
}

func (c *ActionList) Update(msg tea.Msg) (*ActionList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if c.list.SelectedItem() == nil {
				return c, nil
			}
			selectedAction := c.list.SelectedItem().(commands.ScriptAction)
			c.Hide()
			return c, utils.SendMsg(selectedAction)
		}
	}

	var cmds []tea.Cmd

	t, cmd := c.textInput.Update(msg)
	c.textInput = &t
	cmds = append(cmds, cmd)

	l, cmd := c.list.Update(msg)
	c.list = &l
	cmds = append(cmds, cmd)

	return c, cmd
}

func (c *ActionList) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.list.View(), c.footerView())
}
