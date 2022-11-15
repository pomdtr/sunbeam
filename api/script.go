package api

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pomdtr/sunbeam/utils"
)

type Script struct {
	Mode    string     `json:"mode" yaml:"mode"`
	Params  []FormItem `json:"params" yaml:"params"`
	Title   string     `json:"title" yaml:"title"`
	Command string     `json:"command" yaml:"command"`
}

type FormItem struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	Required bool   `json:"required"`
	Default  any    `json:"default"`

	// textitem, textarea
	Secure      bool   `json:"secure"`
	Placeholder string `json:"placeholder"`

	// dropdown
	Data []DropDownItem `json:"data"`

	// checkbox
	Label             string `json:"label"`
	TrueSubstitution  string `json:"trueSubstitution"`
	FalseSubstitution string `json:"falseSubstitution"`
}

type DropDownItem struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

func (s Script) CheckMissingParams(inputParams map[string]any) []FormItem {
	missing := make([]FormItem, 0)
	for _, input := range s.Params {
		if _, ok := inputParams[input.Name]; !ok {
			missing = append(missing, input)
		}
	}
	return missing
}

func (s Script) Cmd(params map[string]any) (*exec.Cmd, error) {
	var err error
	inputs := make(map[string]string)
	for _, formInput := range s.Params {
		value, ok := params[formInput.Name]
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

	rendered, err := utils.RenderString(s.Command, inputs)
	if err != nil {
		return nil, err
	}
	return exec.Command("sh", "-c", rendered), nil
}

type ListItem struct {
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Accessories []string `json:"accessories"`
	Actions     []Action `json:"actions"`
}

type Action struct {
	Title    string `json:"title" yaml:"title"`
	Type     string `json:"type" yaml:"type"`
	Shortcut string `json:"shortcut,omitempty" yaml:"shortcut"`

	Content string `json:"content,omitempty" yaml:"content"`

	Url         string `json:"url,omitempty" yaml:"url"`
	Path        string `json:"path,omitempty" yaml:"path"`
	Application string `json:"application,omitempty" yaml:"application"`

	Extension string `json:"extension,omitempty" yaml:"extension"`
	Script    string `json:"script,omitempty" yaml:"script"`
	Command   string `json:"command,omitempty" yaml:"command"`

	OnSuccess string         `json:"onSuccess" yaml:"onSuccess"`
	With      map[string]any `json:"with,omitempty" yaml:"with"`
}

func ParseListItems(output string) (items []ListItem, err error) {
	rows := strings.Split(output, "\n")
	for _, row := range rows {
		if row == "" {
			continue
		}
		var item ListItem
		err = json.Unmarshal([]byte(row), &item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
