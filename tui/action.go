package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/api"
)

type Action struct {
	Title    string
	Msg      tea.Msg
	Shortcut string
}

type CopyMsg struct {
	Content string
}

type OpenMsg struct {
	Url         string
	Application string
}

type ReloadMsg struct {
	Params map[string]any
}

type RunMsg struct {
	Extension string
	Target    string
	Params    map[string]any
}

func (a Action) SendMsg() tea.Msg {
	return a.Msg
}

func NewAction(scriptAction api.ScriptAction) Action {
	title := scriptAction.Title
	switch scriptAction.Type {
	case "open-url":
		if title == "" {
			title = "Open URL"
		}
		return Action{
			Title:    title,
			Shortcut: scriptAction.Shortcut,
			Msg: OpenMsg{
				Url:         scriptAction.Url,
				Application: scriptAction.Application,
			},
		}
	case "open-file":
		if title == "" {
			title = "Open File"
		}
		return Action{
			Title:    title,
			Shortcut: scriptAction.Shortcut,
			Msg: OpenMsg{
				Url:         scriptAction.Path,
				Application: scriptAction.Application,
			},
		}
	case "copy":
		if title == "" {
			title = "Copy to Clipboard"
		}
		return Action{
			Title:    title,
			Shortcut: scriptAction.Shortcut,
			Msg: CopyMsg{
				Content: scriptAction.Content,
			},
		}
	case "reload":
		if title == "" {
			title = "Reload Script"
		}
		return Action{
			Title:    title,
			Shortcut: scriptAction.Shortcut,
			Msg: ReloadMsg{
				Params: scriptAction.Params,
			},
		}
	case "run":
		if title == "" {
			title = "Run Script"
		}
		return Action{
			Title:    title,
			Shortcut: scriptAction.Shortcut,
			Msg: RunMsg{
				Target: scriptAction.Target,
				Params: scriptAction.Params,
			},
		}
	default:
		return Action{
			Title:    "Unknown",
			Msg:      fmt.Errorf("Unknown action type: %s", scriptAction.Type),
			Shortcut: scriptAction.Shortcut,
		}
	}
}
