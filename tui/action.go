package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/api"
)

type Action struct {
	Msg      tea.Msg
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

type ReloadMsg struct {
	Params map[string]any
}

type PushMsg struct {
	Extension string
	Page      string
	Params    map[string]any
}

type RunMsg struct {
	Extension string
	Script    string
	Params    map[string]any
}

func (a Action) Cmd() tea.Msg {
	return a.Msg
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
	case "copy":
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
		if scriptAction.Title == "" {
			scriptAction.Title = "Open Page"
		}
		msg = PushMsg{
			Extension: scriptAction.Extension,
			Page:      scriptAction.Page,
			Params:    scriptAction.With,
		}
	case "run-script":
		if scriptAction.Title == "" {
			scriptAction.Title = "Run Script"
		}
		msg = RunMsg{
			Extension: scriptAction.Extension,
			Script:    scriptAction.Script,
			Params:    scriptAction.With,
		}
	default:
		scriptAction.Title = "Unknown"
		msg = fmt.Errorf("unknown action type: %s", scriptAction.Type)
	}

	return Action{
		Msg:      msg,
		Title:    scriptAction.Title,
		Shortcut: scriptAction.Shortcut,
	}
}

type ActionList struct {
	filter *Filter
	footer *Footer

	actions []Action
	Shown   bool
}

func NewActionList() *ActionList {
	filter := NewFilter()
	filter.DrawLines = true
	footer := NewFooter("Actions")
	footer.SetBindings(
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("↩", "Select Action")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("⇥", "Hide Actions")),
	)

	return &ActionList{
		filter: filter,
		footer: footer,
	}
}

func (al ActionList) headerView() string {
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, "   ", al.filter.TextInput.View())

	line := strings.Repeat("─", al.footer.Width)
	return lipgloss.JoinVertical(lipgloss.Left, headerRow, line)
}

func (al *ActionList) SetSize(w, h int) {
	availableHeight := h - lipgloss.Height(al.headerView()) - lipgloss.Height(al.footer.View())

	al.filter.SetSize(w, availableHeight)
	al.footer.Width = w
}

func (al *ActionList) Hide() {
	al.Shown = false
}

func (al *ActionList) Show() {
	al.Shown = true
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
		if msg.String() == "tab" {
			al.Shown = !al.Shown
		}

		if msg.String() == "enter" && al.Shown {
			selectedItem := al.filter.Selection()
			if selectedItem == nil {
				break
			}
			listItem, _ := selectedItem.(ListItem)
			al.Hide()
			return &al, listItem.Actions[0].Cmd
		}
		for _, action := range al.actions {
			if key.Matches(msg, action.Binding()) {
				al.Hide()
				return &al, action.Cmd
			}
		}
	}

	var cmd tea.Cmd
	if al.Shown {
		al.filter, cmd = al.filter.Update(msg)
	}
	return &al, cmd
}

func (al ActionList) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		al.headerView(),
		al.filter.View(),
		al.footer.View(),
	)
}
