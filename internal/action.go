package internal

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/browser"
	"github.com/pomdtr/sunbeam/pkg"
)

type ActionList struct {
	actions []pkg.Action
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
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("↩", "Confirm")),
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
			return al, func() tea.Msg {
				selectedItem := al.filter.Selection()
				if selectedItem == nil {
					return nil
				}
				listItem, _ := selectedItem.(ListItem)
				al.Blur()

				if len(listItem.Actions) == 0 {
					return nil
				}

				return listItem.Actions[0]
			}
		default:
			for _, action := range al.actions {
				if msg.String() == fmt.Sprintf("alt+%s", action.Key) {
					return al, func() tea.Msg {
						al.Blur()
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

func runAction(action pkg.Action, extension Extension) tea.Cmd {
	return func() tea.Msg {
		switch action.Type {
		case pkg.ActionTypeCopy:
			if err := clipboard.WriteAll(action.Text); err != nil {
				return fmt.Errorf("could not copy to clipboard: %s", action.Text)
			}

			return ExitMsg{}
		case pkg.ActionTypeOpen:
			if err := browser.OpenURL(action.Url); err != nil {
				return fmt.Errorf("could not open url: %s", action.Url)
			}

			return ExitMsg{}
		case pkg.ActionTypeRun:
			_, err := extension.Run(action.Command.Name, pkg.CommandInput{
				Params: action.Command.Params,
			})
			if err != nil {
				return err
			}

			return ExitMsg{}
		case pkg.ActionTypePush:
			page, err := CommandToPage(extension, pkg.CommandRef{
				Name:   action.Command.Name,
				Params: action.Command.Params,
			})
			if err != nil {
				return err
			}

			return PushPageMsg{Page: page}
		}

		return nil
	}
}
