package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/api"
	"github.com/skratchdot/open-golang/open"
)

type Action struct {
	Title string
	ActionRunner
	Shortcut string
}

type ActionRunner interface {
	Exec() tea.Cmd
}

func NewAction(extension string, scriptAction api.ScriptAction) Action {
	title := scriptAction.Title
	if title == "" {
		title = strings.Title(scriptAction.Type)
	}
	var runner ActionRunner
	switch scriptAction.Type {
	case "open-url":
		runner = OpenUrlRunner{
			url:         scriptAction.Url,
			application: scriptAction.Application,
		}
	case "open-file":
		runner = OpenFileAction{
			path:        scriptAction.Path,
			application: scriptAction.Application,
		}
	case "copy":
		runner = CopyAction{
			content: scriptAction.Content,
		}
	case "run":
		runner = RunAction{
			extension: extension,
			target:    scriptAction.Target,
			params:    scriptAction.Params,
		}
	case "reload":
		runner = ReloadAction{
			params: scriptAction.Params,
		}
	default:
		runner = UnknownAction{}
	}

	return Action{
		Shortcut:     scriptAction.Shortcut,
		ActionRunner: runner,
	}
}

type CopyAction struct {
	content string
}

func (c CopyAction) Exec() tea.Cmd {
	err := clipboard.WriteAll(c.content)
	if err != nil {
		return NewErrorCmd(fmt.Errorf("failed to copy %s to clipboard", err))
	}
	return tea.Quit
}

type RunAction struct {
	extension string
	target    string
	params    map[string]string
}

func (p RunAction) Exec() tea.Cmd {
	command, ok := api.Sunbeam.GetScript(p.extension, p.target)
	if !ok {
		return NewErrorCmd(fmt.Errorf("Unable to find command %s.%s", p.extension, p.target))
	}

	return NewPushCmd(NewRunContainer(command, p.params))
}

type ReloadAction struct {
	params map[string]string
}

func (r ReloadAction) Exec() tea.Cmd {
	input := api.NewScriptInput(r.params)
	return NewReloadCmd(input)
}

type OpenUrlRunner struct {
	application string
	url         string
}

func (o OpenUrlRunner) Exec() tea.Cmd {
	var err error
	if o.application != "" {
		err = open.RunWith(o.url, o.application)
	} else {
		err = open.Run(o.url)
	}

	if err != nil {
		return NewErrorCmd(fmt.Errorf("Unable to open url: %s", err))
	}
	return tea.Quit
}

type OpenFileAction struct {
	application string
	path        string
}

func (o OpenFileAction) Exec() tea.Cmd {
	var err error
	if o.application != "" {
		err = open.RunWith(o.path, o.application)
	} else {
		err = open.Run(o.path)
	}

	if err != nil {
		return NewErrorCmd(fmt.Errorf("Unable to open url: %s", err))
	}
	return tea.Quit
}

type UnknownAction struct {
	actionType string
}

func (u UnknownAction) Exec() tea.Cmd {
	return NewErrorCmd(fmt.Errorf("Unknown action type: %s", u.actionType))
}
