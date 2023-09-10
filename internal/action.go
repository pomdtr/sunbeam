package internal

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/pkg"
)

type ActionList struct {
	actions   []pkg.Action
	header    Header
	filter    Filter
	runAction func(pkg.Action) tea.Cmd
	footer    Footer
}

func NewActionList(runAction func(pkg.Action) tea.Cmd) ActionList {
	filter := NewFilter()
	filter.DrawLines = true

	header := NewHeader()
	footer := NewFooter("Actions")
	footer.SetBindings(
		key.NewBinding(key.WithKeys(""), key.WithHelp("↩", "Confirm")),
	)

	return ActionList{
		runAction: runAction,
		header:    header,
		filter:    filter,
		footer:    footer,
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

func (al *ActionList) SetActions(actions ...pkg.Action) {
	al.actions = actions
	filterItems := make([]FilterItem, len(actions))
	for i, action := range actions {
		var subtitle string
		if i == 0 {
			subtitle = "enter"
		} else if action.Key != "" {
			subtitle = fmt.Sprintf("alt+%s", action.Key)
		}

		filterItems[i] = ListItem(
			pkg.ListItem{
				Title:    action.Title,
				Subtitle: subtitle,
				Actions:  []pkg.Action{action},
			},
		)
	}

	al.filter.SetItems(filterItems)
	al.filter.FilterItems(al.header.Value())
}

func (al ActionList) Update(msg tea.Msg) (ActionList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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
		case "tab":
			if !al.Focused() {
				return al, nil
			}

			al.filter.CursorDown()
		case "shift+tab":
			if !al.Focused() {
				return al, nil
			}

			al.filter.CursorUp()
		case "enter":
			selectedItem := al.filter.Selection()
			if selectedItem == nil {
				return al, nil
			}
			listItem, _ := selectedItem.(ListItem)
			al.Blur()

			if len(listItem.Actions) == 0 {
				return al, nil
			}

			return al, al.runAction(listItem.Actions[0])
		default:
			for _, action := range al.actions {
				if msg.String() == fmt.Sprintf("alt+%s", action.Key) {
					return al, al.runAction(action)
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

type RunMsg struct {
	Command string
	Params  map[string]any
	Exit    bool
}
