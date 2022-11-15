package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
)

type Action struct {
	Cmd      tea.Cmd
	Shortcut string
	Title    string
}

func (a Action) Binding() key.Binding {
	prettyKey := strings.ReplaceAll(a.Shortcut, "ctrl", "⌃")
	prettyKey = strings.ReplaceAll(prettyKey, "alt", "⌥")
	prettyKey = strings.ReplaceAll(prettyKey, "shift", "⇧")
	prettyKey = strings.ReplaceAll(prettyKey, "cmd", "⌘")
	prettyKey = strings.ReplaceAll(prettyKey, "enter", "↩")
	return key.NewBinding(key.WithKeys(a.Shortcut), key.WithHelp(prettyKey, a.Title))
}

type CopyMsg struct {
	Content string
}

type OpenUrlMsg struct {
	Url         string
	Application string
}

type OpenFileMessage struct {
	Path        string
	Application string
}

func NewReloadCmd(params map[string]any) tea.Cmd {
	return func() tea.Msg {
		return ReloadMsg{
			Params: params,
		}
	}
}

type ReloadMsg struct {
	Params map[string]any
}

type RunMsg struct {
	Extension string
	Script    string
	Params    map[string]any
}

type ExecMsg struct {
	Command   *exec.Cmd
	OnSuccess string
}

func NewAction(scriptAction api.Action) Action {
	var msg tea.Msg
	switch scriptAction.Type {
	case "open-url":
		if scriptAction.Title == "" {
			scriptAction.Title = "Open URL"
		}
		msg = OpenUrlMsg{
			Url: scriptAction.Url,
		}
	case "open-file":
		if scriptAction.Title == "" {
			scriptAction.Title = "Open File"
		}
		msg = OpenFileMessage{
			Path:        scriptAction.Path,
			Application: scriptAction.Application,
		}
	case "copy-content":
		if scriptAction.Title == "" {
			scriptAction.Title = "Copy to Clipboard"
		}
		msg = CopyMsg{
			Content: scriptAction.Content,
		}
	case "reload-page":
		if scriptAction.Title == "" {
			scriptAction.Title = "Reload Script"
		}
		msg = ReloadMsg{
			Params: scriptAction.With,
		}
	case "push-page":
		msg = RunMsg{
			Extension: scriptAction.Extension,
			Script:    scriptAction.Script,
			Params:    scriptAction.With,
		}
	case "exec-command":
		msg = ExecMsg{
			Command:   exec.Command("sh", "-c", scriptAction.Command),
			OnSuccess: scriptAction.OnSuccess,
		}
	default:
		scriptAction.Title = "Unknown"
		msg = fmt.Errorf("unknown action type: %s", scriptAction.Type)
	}

	return Action{
		Cmd: func() tea.Msg {
			return msg
		},
		Title:    scriptAction.Title,
		Shortcut: scriptAction.Shortcut,
	}
}

type ActionList struct {
	header Header
	filter *Filter
	footer Footer

	actions []Action
}

func NewActionList() *ActionList {
	filter := NewFilter()
	filter.DrawLines = true

	header := NewHeader()
	footer := NewFooter("Actions")
	footer.SetBindings(
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("↩", "Confirm")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Hide Actions")),
	)

	return &ActionList{
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
		filterItems[i] = ListItem{
			Title:    action.Title,
			Subtitle: action.Shortcut,
			Actions:  []Action{action},
		}
	}
	al.filter.SetItems(filterItems)
}

func (al ActionList) Update(msg tea.Msg) (*ActionList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if !al.Focused() && len(al.actions) > 1 {
				return &al, al.Focus()
			} else if al.Focused() {
				al.header.input.SetValue("")
				al.filter.FilterItems("")
				al.Blur()
				return &al, nil
			}
		case "esc":
			if !al.Focused() {
				return &al, nil
			}

			if al.header.input.Value() != "" {
				al.header.input.SetValue("")
				al.filter.FilterItems("")
				return &al, nil
			}

			al.Blur()
			return &al, nil

		case "enter":
			selectedItem := al.filter.Selection()
			if selectedItem == nil {
				return &al, nil
			}
			listItem, _ := selectedItem.(ListItem)
			al.Blur()
			return &al, listItem.Actions[0].Cmd
		}

		for _, action := range al.actions {
			if key.Matches(msg, action.Binding()) {
				al.Blur()
				return &al, action.Cmd
			}
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	if !al.Focused() {
		return &al, nil
	}

	header, cmd := al.header.Update(msg)
	cmds = append(cmds, cmd)
	if header.Value() != al.header.Value() {
		cmd = al.filter.FilterItems(header.Value())
		cmds = append(cmds, cmd)
	}
	al.header = header

	al.filter, cmd = al.filter.Update(msg)
	cmds = append(cmds, cmd)

	return &al, tea.Batch(cmds...)
}

func (al ActionList) Focused() bool {
	return al.header.input.Focused()

}

func (al *ActionList) Focus() tea.Cmd {
	return al.header.Focus()
}

func (al *ActionList) Blur() {
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
