package tui

import (
	"errors"
	"log"
	"os/exec"
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
	case "launch":
		return NewLaunchAction(title, scriptAction.Shortcut, extension, scriptAction.Target, scriptAction.Params)
	case "reload":
		return NewReloadAction(title, scriptAction.Shortcut, scriptAction.Params)
	case "exec":
		return NewExecAction(title, scriptAction.Shortcut, scriptAction.Command)
	default:
		return NewUnknownAction(scriptAction.Type)
	}
}

type BaseAction struct {
	title string
	key   string
}

func (b BaseAction) Title() string {
	return b.title
}

func (b BaseAction) Shortcut() string {
	return b.key
}

type CopyAction struct {
	BaseAction
	Content string
}

func NewCopyAction(title string, key string, content string) Action {
	return CopyAction{BaseAction: BaseAction{title: title, key: key}, Content: content}
}

func (c CopyAction) Exec() tea.Cmd {
	err := clipboard.WriteAll(c.Content)
	if err != nil {
		return NewErrorCmd("failed to copy %s to clipboard", err)
	}
	return tea.Quit
}

type LaunchAction struct {
	BaseAction
	extension string
	target    string
	params    map[string]string
}

func NewLaunchAction(title string, key string, extensionName string, target string, params map[string]string) Action {
	return LaunchAction{BaseAction: BaseAction{title: title, key: key}, extension: extensionName, target: target, params: params}
}

func (p LaunchAction) Exec() tea.Cmd {
	extension, ok := api.Extensions[p.extension]
	if !ok {
		return NewErrorCmd("unknown extension %s", p.target)
	}
	command, ok := extension.Commands[p.target]
	if !ok {
		return NewErrorCmd("unknown command %s", p.target)
	}
	log.Println("launching", p.extension, p.target, p.params)

	return NewPushCmd(NewCommandContainer(command, p.params))
}

type ReloadAction struct {
	BaseAction
	params map[string]string
}

func NewReloadAction(title string, key string, params map[string]string) Action {
	return ReloadAction{BaseAction: BaseAction{title: title, key: key}, params: params}
}

func (r ReloadAction) Exec() tea.Cmd {
	input := api.NewCommandInput(r.params)
	return NewReloadCmd(input)
}

type ExecAction struct {
	BaseAction
	command string
}

func NewExecAction(title string, key string, command string) Action {
	return ExecAction{BaseAction: BaseAction{title: title, key: key}, command: command}
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

func NewOpenUrlAction(title string, key string, url string, application string) Action {
	return OpenUrlAction{BaseAction: BaseAction{title: title, key: key}, application: application, url: url}
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

func NewOpenFileAction(title string, key string, path string, application string) Action {
	return OpenFileAction{BaseAction: BaseAction{title: title, key: key}, application: application, path: path}
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
