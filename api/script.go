package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/pomdtr/sunbeam/utils"
)

type Script struct {
	Inputs  []FormItem `json:"params"`
	Command string     `json:"command"`
}

type FormItem struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	Required bool   `json:"required"`

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

type ScriptInputs struct {
	Inputs map[string]any
	Query  string
}

func (s Script) CheckMissingParams(inputParams map[string]any) []FormItem {
	missing := make([]FormItem, 0)
	for _, input := range s.Inputs {
		if _, ok := inputParams[input.Name]; !ok {
			log.Println(input.Name, "is missing")
			missing = append(missing, input)
		}
	}
	return missing
}

func (s Script) Run(dir string, params map[string]any) (string, error) {
	var err error
	inputs := make(map[string]string)
	for _, formInput := range s.Inputs {
		value, ok := params[formInput.Name]
		if !ok {
			return "", fmt.Errorf("missing param %s", formInput.Name)
		}
		if formInput.Type == "checkbox" {
			value, ok := value.(bool)
			if !ok {
				return "", fmt.Errorf("invalid type for param %s", formInput.Name)
			}
			if value {
				inputs[formInput.Name] = shellescape.Quote(formInput.TrueSubstitution)
			} else if formInput.FalseSubstitution != "" {
				inputs[formInput.Name] = shellescape.Quote(formInput.FalseSubstitution)
			} else {
				inputs[formInput.Name] = ""
			}
		} else {
			value, ok := value.(string)
			if !ok {
				return "", fmt.Errorf("param %s is not a string", formInput.Name)
			}
			inputs[formInput.Name] = shellescape.Quote(value)
		}
	}

	rendered, err := utils.RenderString(s.Command, inputs)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("sh", "-c", rendered)
	cmd.Dir = dir
	log.Printf("Running '%s' in %s", rendered, cmd.Dir)

	var outbuf, errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error while running command: %s", errbuf.String())
	}
	return outbuf.String(), nil
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
	Page      string `json:"page,omitempty" yaml:"page"`
	Script    string `json:"script,omitempty" yaml:"script"`

	With map[string]any `json:"with,omitempty" yaml:"with"`
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
