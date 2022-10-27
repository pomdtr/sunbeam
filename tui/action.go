package tui

import (
	"errors"
	"os/exec"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/api"
	"github.com/skratchdot/open-golang/open"
)

type Action interface {
	Title() string
	Exec() tea.Cmd
	Keybind() string
}

func NewAction(scriptAction api.ScriptAction) Action {
	title := scriptAction.Title
	switch scriptAction.Type {
	case "open-url":
		if title == "" {
			title = "Open URL"
		}
		return NewOpenUrlAction(title, scriptAction.Keybind, scriptAction.Url, scriptAction.Application)
	case "open-file":
		if title == "" {
			title = "Open File"
		}
		return NewOpenFileAction(title, scriptAction.Keybind, scriptAction.Url, scriptAction.Application)
	case "copy":
		if title == "" {
			title = "Copy to Clibpoard"
		}
		return NewCopyAction(title, scriptAction.Keybind, scriptAction.Content)
	case "push":
		if title == "" {
			title = "Push"
		}
		return NewPushAction(title, scriptAction.Keybind, scriptAction.Target, scriptAction.Params)
	case "reload":
		if title == "" {
			title = "Reload"
		}
		return NewReloadAction(title, scriptAction.Keybind, scriptAction.Params)
	case "exec":
		if title == "" {
			title = "Run"
		}
		return NewExecAction(title, scriptAction.Keybind, scriptAction.Command)
	default:
		return NewUnknownAction(scriptAction.Type)
	}
}

type BaseAction struct {
	title   string
	keybind string
}

func (b BaseAction) Title() string {
	return b.title
}

func (b BaseAction) Keybind() string {
	return b.keybind
}

type CopyAction struct {
	BaseAction
	Content string
}

func NewCopyAction(title string, keybind string, content string) Action {
	return CopyAction{BaseAction: BaseAction{title: title, keybind: keybind}, Content: content}
}

func (c CopyAction) Exec() tea.Cmd {
	err := clipboard.WriteAll(c.Content)
	if err != nil {
		return NewErrorCmd("failed to copy %s to clipboard", err)
	}
	return tea.Quit
}

type PushAction struct {
	BaseAction
	target string
	params map[string]string
}

func NewPushAction(title string, keybind string, target string, params map[string]string) Action {
	return PushAction{BaseAction: BaseAction{title: title, keybind: keybind}, target: target, params: params}
}

func (p PushAction) Exec() tea.Cmd {
	command, ok := api.GetSunbeamCommand(p.target)
	if !ok {
		return NewErrorCmd("unknown command %s", p.target)
	}

	return NewPushCmd(NewCommandContainer(command, p.params))
}

type ReloadAction struct {
	BaseAction
	params map[string]string
}

func NewReloadAction(title string, keybind string, params map[string]string) Action {
	return ReloadAction{BaseAction: BaseAction{title: title, keybind: keybind}, params: params}
}

func (r ReloadAction) Exec() tea.Cmd {
	input := api.NewCommandInput(r.params)
	return NewReloadCmd(input)
}

type ExecAction struct {
	BaseAction
	command string
}

func NewExecAction(title string, keybind string, command string) Action {
	return ExecAction{BaseAction: BaseAction{title: title, keybind: keybind}, command: command}
}

func (e ExecAction) Exec() tea.Cmd {
	cmd := exec.Command("sh", "-c", e.command)
	_, err := cmd.Output()
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return NewErrorCmd("Unable to run cmd: %s", exitError.Stderr)
	}
	return tea.Quit
}

type OpenUrlAction struct {
	BaseAction
	application string
	url         string
}

func NewOpenUrlAction(title string, keybind string, url string, application string) Action {
	return OpenUrlAction{BaseAction: BaseAction{title: title, keybind: keybind}, application: application, url: url}
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
	BaseAction
	application string
	path        string
}

func NewOpenFileAction(title string, keybind string, path string, application string) Action {
	return OpenFileAction{BaseAction: BaseAction{title: title, keybind: keybind}, application: application, path: path}
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
	BaseAction
}

func NewUnknownAction(actionType string) Action {
	return UnknownAction{BaseAction: BaseAction{title: "Unknown"}}
}

func (u UnknownAction) Exec() tea.Cmd {
	return nil
}
