package app

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
	"github.com/pomdtr/sunbeam/utils"
)

type Script struct {
	Command   string        `json:"command" yaml:"command"`
	OnSuccess string        `json:"onSuccess" yaml:"onSuccess"`
	Cwd       string        `json:"cwd" yaml:"cwd"`
	Inputs    []ScriptInput `json:"inputs" yaml:"inputs"`
	Page      Page          `json:"page" yaml:"page"`
}

type Page struct {
	Type        string `json:"type" yaml:"type"`
	Title       string `json:"title" yaml:"title"`
	Mode        string `json:"mode" yaml:"mode"`
	ShowPreview bool   `json:"showPreview" yaml:"showPreview"`
}

type ScriptInput struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	Required bool   `json:"required"`
	Default  any    `json:"default"`

	// textitem, textarea
	Placeholder string `json:"placeholder"`

	// dropdown
	Data []struct {
		Title string `json:"title"`
		Value string `json:"value"`
	} `json:"data"`

	// checkbox
	Label             string `json:"label"`
	TrueSubstitution  string `json:"trueSubstitution"`
	FalseSubstitution string `json:"falseSubstitution"`
}

func (s Script) CheckMissingParams(inputParams map[string]any) []ScriptInput {
	missing := make([]ScriptInput, 0)
	for _, input := range s.Inputs {
		value, ok := inputParams[input.Name]
		if !ok {
			missing = append(missing, input)
		}
		if input.Type != "checkbox" && value == "" {
			missing = append(missing, input)
		}
	}
	return missing
}

type CommandParams struct {
	With  map[string]any
	Query string
}

func (s Script) Cmd(params CommandParams) (*exec.Cmd, error) {
	var err error

	inputs := make(map[string]string)
	for _, formInput := range s.Inputs {
		value, ok := params.With[formInput.Name]
		if !ok {
			return nil, fmt.Errorf("missing param %s", formInput.Name)
		}
		if formInput.Type == "checkbox" {
			value, ok := value.(bool)
			if !ok {
				return nil, fmt.Errorf("invalid type for param %s", formInput.Name)
			}
			if value {
				inputs[formInput.Name] = formInput.TrueSubstitution
			} else if formInput.FalseSubstitution != "" {
				inputs[formInput.Name] = formInput.FalseSubstitution
			} else {
				inputs[formInput.Name] = ""
			}
		} else {
			value, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("param %s is not a string", formInput.Name)
			}
			inputs[formInput.Name] = value
		}
	}

	funcMap := template.FuncMap{
		"query": func() string {
			return shellescape.Quote(params.Query)
		},
	}

	for key, value := range inputs {
		funcMap[key] = func() string {
			return shellescape.Quote(value)
		}
	}

	rendered, err := utils.RenderString(s.Command, funcMap)
	if err != nil {
		return nil, err
	}
	return exec.Command("sh", "-c", rendered), nil
}

type ScriptItem struct {
	Id          string         `json:"id"`
	Title       string         `json:"title"`
	Subtitle    string         `json:"subtitle"`
	Preview     string         `json:"preview"`
	PreviewCmd  string         `json:"previewCmd"`
	Accessories []string       `json:"accessories"`
	Actions     []ScriptAction `json:"actions"`
}

func (li ScriptItem) PreviewCommand() *exec.Cmd {
	if li.PreviewCmd == "" {
		return nil
	}
	return exec.Command("sh", "-c", li.PreviewCmd)
}

type ScriptAction struct {
	Title    string `json:"title" yaml:"title"`
	Type     string `json:"type" yaml:"type"`
	Shortcut string `json:"shortcut,omitempty" yaml:"shortcut"`

	Text string `json:"text,omitempty" yaml:"textfield"`

	Url         string `json:"url,omitempty" yaml:"url"`
	Path        string `json:"path,omitempty" yaml:"path"`
	Application string `json:"application,omitempty" yaml:"application"`

	Extension string `json:"extension,omitempty" yaml:"extension"`
	Script    string `json:"script,omitempty" yaml:"script"`
	Command   string `json:"command,omitempty" yaml:"command"`

	OnSuccess string         `json:"onSuccess" yaml:"onSuccess"`
	With      map[string]any `json:"with,omitempty" yaml:"with"`
}

func ParseAction(output string) (action ScriptAction, err error) {
	err = json.Unmarshal([]byte(output), &action)
	return action, err
}

func ParseListItems(output string) (items []ScriptItem, err error) {
	rows := strings.Split(output, "\n")
	for _, row := range rows {
		if row == "" {
			continue
		}
		var item ScriptItem
		err = json.Unmarshal([]byte(row), &item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
