package internal

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/types"
)

type ActionList struct {
	actions []types.Action
	header  Header
	filter  Filter
	footer  Footer
}

func NewActionList() ActionList {
	filter := NewFilter()
	filter.DrawLines = true

	header := NewHeader()
	footer := NewFooter("Actions")
	footer.SetBindings(
		key.NewBinding(key.WithKeys(""), key.WithHelp("↩", "Confirm")),
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

func (al *ActionList) SetActions(actions ...types.Action) {
	al.actions = actions
	filterItems := make([]FilterItem, len(actions))
	for i, action := range actions {
		var subtitle string
		if i == 0 {
			subtitle = "enter"
		} else if i == 1 {
			subtitle = "alt+enter"
		} else if action.Key != "" {
			subtitle = fmt.Sprintf("alt+%s", action.Key)
		}

		filterItems[i] = ListItem{
			ListItem: types.ListItem{
				Title:    action.Title,
				Subtitle: subtitle,
				Actions:  []types.Action{action},
			},
		}
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
		case "enter":
			selectedItem := al.filter.Selection()
			if selectedItem == nil {
				return al, nil
			}
			listItem, _ := selectedItem.(ListItem)
			al.Blur()

			return al, func() tea.Msg {
				if len(listItem.Actions) == 0 {
					return nil
				}

				return listItem.Actions[0]
			}
		case "ctrl+y":
			selectedItem := al.filter.Selection()
			if selectedItem == nil {
				return al, nil
			}
			listItem, _ := selectedItem.(ListItem)
			al.Blur()

			return al, func() tea.Msg {
				if len(listItem.Actions) == 0 {
					return nil
				}
				action := listItem.Actions[0]
				var content string
				switch action.Type {
				case types.CopyAction:
					content = action.Text
				case types.OpenAction:
					content = action.Target
				case types.RunAction:
					content = action.Command.Cmd().String()
				case types.PushAction:
					if action.Page.Command != nil {
						content = action.Command.Cmd().String()
					} else {
						content = action.Page.Text
					}
				}

				return clipboard.WriteAll(content)
			}
		default:
			for _, action := range al.actions {
				if msg.String() == fmt.Sprintf("alt+%s", action.Key) {
					return al, func() tea.Msg {
						return action
					}
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
