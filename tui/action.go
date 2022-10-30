package tui

import (
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/api"
	"github.com/skratchdot/open-golang/open"
)

type Action interface {
	Title() string
	Exec() tea.Cmd
	Shortcut() string
}

func NewAction(extension string, scriptAction api.ScriptAction) Action {
	title := scriptAction.Title
	if title == "" {
		title = strings.Title(scriptAction.Type)
	}
	switch scriptAction.Type {
	case "open-url":
		return NewOpenUrlAction(title, scriptAction.Shortcut, scriptAction.Url, scriptAction.Application)
	case "open-file":
		return NewOpenFileAction(title, scriptAction.Shortcut, scriptAction.Url, scriptAction.Application)
	case "copy":
		return NewCopyAction(title, scriptAction.Shortcut, scriptAction.Content)
	case "run":
		return NewRunAction(title, scriptAction.Shortcut, extension, scriptAction.Target, scriptAction.Params)
	case "reload":
		return NewReloadAction(title, scriptAction.Shortcut, scriptAction.Params)
	default:
		return NewUnknownAction(scriptAction.Type)
	}
}

type baseAction struct {
	title string
	key   string
}

func (b baseAction) Title() string {
	return b.title
}

func (b baseAction) Shortcut() string {
	return b.key
}

type CopyAction struct {
	baseAction
	Content string
}

func NewCopyAction(title string, key string, content string) Action {
	return CopyAction{baseAction: baseAction{title: title, key: key}, Content: content}
}

func (c CopyAction) Exec() tea.Cmd {
	err := clipboard.WriteAll(c.Content)
	if err != nil {
		return NewErrorCmd("failed to copy %s to clipboard", err)
	}
	return tea.Quit
}

type RunAction struct {
	baseAction
	extension string
	target    string
	params    map[string]string
}

func NewRunAction(title string, key string, extensionName string, target string, params map[string]string) Action {
	return RunAction{baseAction: baseAction{title: title, key: key}, extension: extensionName, target: target, params: params}
}

func (p RunAction) Exec() tea.Cmd {
	command, ok := api.Sunbeam.GetScript(p.extension, p.target)
	if !ok {
		return NewErrorCmd("Unable to find command %s.%s", p.extension, p.target)
	}

	return NewPushCmd(NewRunContainer(command, p.params))
}

type ReloadAction struct {
	baseAction
	params map[string]string
}

func NewReloadAction(title string, key string, params map[string]string) Action {
	return ReloadAction{baseAction: baseAction{title: title, key: key}, params: params}
}

func (r ReloadAction) Exec() tea.Cmd {
	input := api.NewScriptInput(r.params)
	return NewReloadCmd(input)
}

type OpenUrlAction struct {
	baseAction
	application string
	url         string
}

func NewOpenUrlAction(title string, key string, url string, application string) Action {
	return OpenUrlAction{baseAction: baseAction{title: title, key: key}, application: application, url: url}
}

func (o OpenUrlAction) Exec() tea.Cmd {
	var err error
	if o.application != "" {
		err = open.RunWith(o.url, o.application)
	} else {
		err = open.Run(o.url)
	}

	if err != nil {
		return NewErrorCmd("Unable to open url: %s", err)
	}
	return tea.Quit
}

type OpenFileAction struct {
	baseAction
	application string
	path        string
}

func NewOpenFileAction(title string, key string, path string, application string) Action {
	return OpenFileAction{baseAction: baseAction{title: title, key: key}, application: application, path: path}
}

func (o OpenFileAction) Exec() tea.Cmd {
	var err error
	if o.application != "" {
		err = open.RunWith(o.path, o.application)
	} else {
		err = open.Run(o.path)
	}

	if err != nil {
		return NewErrorCmd("Unable to open url: %s", err)
	}
	return tea.Quit
}

type UnknownAction struct {
	baseAction
}

func NewUnknownAction(actionType string) Action {
	return UnknownAction{baseAction: baseAction{title: "Unknown"}}
}

func (u UnknownAction) Exec() tea.Cmd {
	return nil
}
