package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/scripts"
)

type Action struct {
	Cmd      tea.Cmd
	Title    string
	Shortcut string
}

type CopyTextMsg struct {
	Text string
}

type OpenMsg struct {
	Target string
}

type ReloadPageMsg struct {
}

type PushPageMsg struct {
	Fields []scripts.Field
}

type RunCommandMsg struct {
	Fields []scripts.Field
}

func NewAction(scriptAction scripts.Action) Action {
	var cmd tea.Cmd
	switch scriptAction.Type {
	case "copy":
		if scriptAction.Title == "" {
			scriptAction.Title = "Copy to Clipboard"
		}
		cmd = func() tea.Msg {
			return CopyTextMsg{Text: scriptAction.Target}
		}
	case "reload":
		if scriptAction.Title == "" {
			scriptAction.Title = "Reload Page"
		}
		cmd = func() tea.Msg {
			return ReloadPageMsg{}
		}
	case "run":
		cmd = func() tea.Msg {
			return RunCommandMsg{
				Fields: scriptAction.Command,
			}
		}
	case "open":
		if scriptAction.Title == "" {
			scriptAction.Title = "Open"
		}
		cmd = func() tea.Msg {
			return OpenMsg{Target: scriptAction.Target}
		}
	case "push":
		cmd = func() tea.Msg {
			return PushPageMsg{Fields: scriptAction.Command}
		}
	default:
		scriptAction.Title = "Unknown"
		cmd = func() tea.Msg {
			return fmt.Errorf("unknown action type: %s", scriptAction.Type)
		}
	}

	return Action{
		Cmd:      cmd,
		Shortcut: scriptAction.Shortcut,
		Title:    scriptAction.Title,
	}
}

type ActionList struct {
	header Header
	filter Filter
	footer Footer

	globalActions []Action
	actions       []Action
}

func NewActionList(actions []Action) ActionList {
	filter := NewFilter()
	filter.DrawLines = true

	header := NewHeader()
	footer := NewFooter("Actions")
	footer.SetBindings(
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("↩", "Confirm")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("⎋", "Hide Actions")),
	)

	return ActionList{
		header:        header,
		filter:        filter,
		globalActions: actions,
		footer:        footer,
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
	filterItems := make([]FilterItem, len(actions)+len(al.globalActions))
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

	for i, action := range al.globalActions {
		filterItems[i+len(actions)] = ListItem{
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
