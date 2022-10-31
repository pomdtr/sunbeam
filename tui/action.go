package tui

import (
	"fmt"
	"strings"

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
	Params map[string]string
}

type RunMsg struct {
	Extension string
	Target    string
	Params    map[string]string
}

func (a Action) SendMsg() tea.Msg {
	return a.Msg
}

func NewAction(scriptAction api.ScriptAction) Action {
	title := scriptAction.Title
	if title == "" {
		title = strings.Title(scriptAction.Type)
	}
	switch scriptAction.Type {
	case "open-url":
		return Action{
			Title:    title,
			Shortcut: scriptAction.Shortcut,
			Msg: OpenMsg{
				Url:         scriptAction.Url,
				Application: scriptAction.Application,
			},
		}
	case "open-file":
		return Action{
			Title:    title,
			Shortcut: scriptAction.Shortcut,
			Msg: OpenMsg{
				Url:         scriptAction.Path,
				Application: scriptAction.Application,
			},
		}
	case "copy":
		return Action{
			Title:    title,
			Shortcut: scriptAction.Shortcut,
			Msg: CopyMsg{
				Content: scriptAction.Content,
			},
		}
	case "reload":
		return Action{
			Title:    title,
			Shortcut: scriptAction.Shortcut,
			Msg: ReloadMsg{
				Params: scriptAction.Params,
			},
		}
	case "run":
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
