package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/schemas"
)

type Action struct {
	Cmd      tea.Cmd
	Title    string
	Shortcut string
}

type CopyTextMsg struct {
	Action schemas.Action
}

type OpenMsg struct {
	Action schemas.Action
}

type ReloadPageMsg struct {
	Action schemas.Action
}

type PushPageMsg struct {
	Action schemas.Action
}

type RunCommandMsg struct {
	Action schemas.Action
}

func NewAction(scriptAction schemas.Action, dir string) Action {
	var msg tea.Msg
	if scriptAction.Dir == "" {
		scriptAction.Dir = dir
	}
	switch scriptAction.Type {
	case "copy":
		if scriptAction.Title == "" {
			scriptAction.Title = "Copy"
		}
		msg = CopyTextMsg{Action: scriptAction}
	case "run":
		if scriptAction.Title == "" {
			scriptAction.Title = "Run"
		}

		msg = RunCommandMsg{
			Action: scriptAction,
		}
	case "open":
		if scriptAction.Title == "" {
			scriptAction.Title = "Open"
		}

		msg = OpenMsg{Action: scriptAction}
	case "read":
		if scriptAction.Title == "" {
			scriptAction.Title = "Push"
		}
		msg = PushPageMsg{Action: scriptAction}
	default:
		scriptAction.Title = "Unknown"
		msg = fmt.Errorf("unknown action type: %s", scriptAction.Type)
	}

	return Action{
		Cmd:      func() tea.Msg { return msg },
		Shortcut: scriptAction.Shortcut,
		Title:    scriptAction.Title,
	}
}

type ActionList struct {
	header Header
	filter Filter
	footer Footer

	actions []Action
}

func NewActionList() ActionList {
	filter := NewFilter()
	filter.DrawLines = true

	header := NewHeader()
	footer := NewFooter("Actions")
	footer.SetBindings(
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("↩", "Confirm")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("⎋", "Hide Actions")),
	)

	return ActionList{
		header: header,
		filter: filter,
		footer: footer,
	}
}

func (al *ActionList) SetSize(w, h int) {
	availableHeight := h - lipgloss.Height(al.header.View()) - lipgloss.Height(al.footer.View())

	al.filter.SetSize(w, availableHeight)
	al.footer.Width = w
	al.header.Width = w
}

func (al *ActionList) SetTitle(title string) {
	al.footer.title = title
}

func (al *ActionList) SetActions(actions ...Action) {
	al.actions = actions
	filterItems := make([]FilterItem, len(actions))
	for i, action := range actions {
		if i == 0 {
			action.Shortcut = "↩"
		}
		filterItems[i] = ListItem{
			Title:    action.Title,
			Subtitle: action.Shortcut,
			Actions:  []Action{action},
		}
	}

	al.filter.SetItems(filterItems)
	al.filter.FilterItems(al.header.Value())
}

func (al ActionList) Update(msg tea.Msg) (ActionList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			if !al.Focused() {
				return al, al.Focus()
			} else if al.Focused() {
				if msg.String() == "tab" {
					al.filter.CursorDown()
				} else {
					al.filter.CursorUp()
				}
			}
		case "esc":
			if !al.Focused() {
				return al, nil
			}

			if al.header.input.Value() != "" {
				al.Clear()
			} else {
				al.Blur()
			}

			return al, nil
		case "enter":
			selectedItem := al.filter.Selection()
			if selectedItem == nil {
				return al, nil
			}
			listItem, _ := selectedItem.(ListItem)
			al.Blur()
			return al, listItem.Actions[0].Cmd
		default:
			for _, action := range al.actions {
				if action.Shortcut == msg.String() {
					al.Blur()
					return al, action.Cmd
				}
			}
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	if !al.Focused() {
		return al, nil
	}

	header, cmd := al.header.Update(msg)
	cmds = append(cmds, cmd)
	if header.Value() != al.header.Value() {
		al.filter.FilterItems(header.Value())
	}
	al.header = header

	al.filter, cmd = al.filter.Update(msg)
	cmds = append(cmds, cmd)

	return al, tea.Batch(cmds...)
}

func (al ActionList) Focused() bool {
	return al.header.input.Focused()

}

func (al *ActionList) Focus() tea.Cmd {
	return al.header.Focus()
}

func (al *ActionList) Clear() {
	al.header.input.SetValue("")
	al.filter.FilterItems("")
	al.filter.cursor = 0
}

func (al *ActionList) Blur() {
	al.Clear()
	al.header.input.Blur()
}

func (al ActionList) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		al.header.View(),
		al.filter.View(),
		al.footer.View(),
	)
}
