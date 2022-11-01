package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
)

type Script struct {
	Output  string     `json:"output"`
	Mode    string     `json:"mode"`
	Format  string     `json:"format"`
	Params  []FormItem `json:"params"`
	Command string     `json:"command"`

	Extension string
	Url       url.URL
	Root      url.URL
}

type FormItem struct {
	Type     string `json:"type"`
	Id       string `json:"id"`
	Title    string `json:"title"`
	Required bool   `json:"required"`

	TextField
	DropDown
	Checkbox
}

type TextField struct {
	Secure      bool   `json:"secure"`
	Placeholder string `json:"placeholder"`
}

type DropDown struct {
	Data []DropDownItem `json:"data"`
}

type Checkbox struct {
	Label             string `json:"label"`
	TrueSubstitution  string `json:"trueSubstitution"`
	FalseSubstitution string `json:"falseSubstitution"`
}

type DropDownItem struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type ScriptInput struct {
	Params map[string]any
	Query  string
}

func NewScriptInput(params map[string]any) ScriptInput {
	if params == nil {
		params = make(map[string]any)
	}
	return ScriptInput{Params: params}
}

func (s Script) CheckMissingParams(inputParams map[string]any) []FormItem {
	missing := make([]FormItem, 0)
	for _, param := range s.Params {
		if _, ok := inputParams[param.Id]; !ok {
			missing = append(missing, param)
		}
	}
	return missing
}

func renderCommand(command string, data map[string]any) (string, error) {
	t, err := template.New("").Parse(command)
	if err != nil {
		return "", err
	}

	out := bytes.Buffer{}
	if err = t.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (s Script) Run(input ScriptInput) (string, error) {
	var err error
	params := make(map[string]string)
	for _, formInput := range s.Params {
		value, ok := input.Params[formInput.Id]
		if !ok {
			return "", fmt.Errorf("missing param %s", formInput.Id)
		}
		if formInput.Type == "checkbox" {
			value, ok := value.(bool)
			if !ok {
				return "", fmt.Errorf("invalid type for param %s", formInput.Id)
			}
			if value {
				params[formInput.Id] = shellescape.Quote(formInput.TrueSubstitution)
			} else if formInput.FalseSubstitution != "" {
				params[formInput.Id] = shellescape.Quote(formInput.FalseSubstitution)
			} else {
				params[formInput.Id] = ""
			}
		} else {
			value, ok := value.(string)
			if !ok {
				return "", fmt.Errorf("param %s is not a string", formInput.Id)
			}
			params[formInput.Id] = shellescape.Quote(value)
		}
	}

	rendered, err := renderCommand(s.Command, map[string]any{
		"params": params,
		"query":  shellescape.Quote(input.Query),
	})
	log.Println("Rendered command:", rendered)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("sh", "-c", rendered)
	cmd.Dir = s.Root.Path

	var outbuf, errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error while running command: %s", errbuf.String())
	}
	return outbuf.String(), nil
}

func (s Script) RemoteRun(input ScriptInput) (string, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	res, err := http.Post(http.MethodPost, s.Url.String(), bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	bytes, _ := io.ReadAll(res.Body)

	return string(bytes), nil
}

type ListItem struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Detail   struct {
		Command string `json:"command"`
		Format  string `json:"format"`
	} `json:"detail"`
	Actions     []ScriptAction `json:"actions"`
	Accessories []string       `json:"accessories"`

	Extension string
}

type ScriptAction struct {
	Title    string `json:"title"`
	Type     string `json:"type"`
	Shortcut string `json:"shortcut"`

	CopyAction
	RunAction
	OpenAction
}

type CopyAction struct {
	Content string `json:"content,omitempty"`
}

type OpenAction struct {
	Url         string `json:"url,omitempty"`
	Path        string `json:"path"`
	Application string `json:"application,omitempty"`
}

type RunAction struct {
	Target    string         `json:"target,omitempty"`
	Extension string         `json:"extension,omitempty"`
	Params    map[string]any `json:"params"`
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
