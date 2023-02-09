package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/pomdtr/sunbeam/utils"
	"gopkg.in/yaml.v3"
	"mvdan.cc/sh/v3/shell"
	"mvdan.cc/sh/v3/syntax"
)

type Command struct {
	Name        string            `json:"name"`
	Exec        string            `json:"exec,omitempty" yaml:"exec,omitempty"`
	Env         map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Params      []Param           `json:"params,omitempty" yaml:"params,omitempty"`
	OnSuccess   string            `json:"onSuccess,omitempty" yaml:"onSuccess,omitempty"`
}

type Env struct {
	Name string `json:"name"`
}

type Arg struct {
	Value any
	Input *FormItem
}

func (i *Arg) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		i.Value = s
		return nil
	}

	var boolean bool
	if err := json.Unmarshal(b, &boolean); err == nil {
		i.Value = b
		return nil
	}

	var input FormItem
	if err := json.Unmarshal(b, &input); err == nil {
		i.Input = &input
		return nil
	}

	return fmt.Errorf("invalid input: %s", b)
}

func (i *Arg) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err == nil {
		i.Value = s
		return nil
	}

	var boolean bool
	if err := node.Decode(&boolean); err == nil {
		i.Value = boolean
		return nil
	}

	var input FormItem
	if err := node.Decode(&input); err == nil {
		i.Input = &input
		return nil
	}

	return fmt.Errorf("invalid input: %s", node.Value)
}

func (i Arg) MarshalYAML() (interface{}, error) {
	if i.Value != nil {
		return i.Value, nil
	}
	return i.Input, nil
}

func (i Arg) MarshalJSON() ([]byte, error) {
	if i.Value != nil {
		return json.Marshal(i.Value)
	}
	return json.Marshal(i.Input)
}

type Param struct {
	Name        string    `json:"name"`
	Env         string    `json:"env,omitempty" yaml:"env,omitempty"`
	Input       *FormItem `json:"input,omitempty" yaml:"input,omitempty"`
	Type        string    `json:"type"`
	Default     any       `json:"default,omitempty" yaml:"default,omitempty"`
	Description string    `json:",omitempty" yaml:",omitempty"`
}

func (p Param) FormItem() *FormItem {
	if p.Input != nil {
		return p.Input
	}

	switch p.Type {
	case "string", "file", "directory":
		return &FormItem{
			Type:        "textfield",
			Title:       p.Name,
			Placeholder: p.Description,
		}
	case "bool":
		return &FormItem{
			Type:  "checkbox",
			Title: p.Name,
			Label: p.Description,
		}
	default:
		return nil
	}
}

type FormItem struct {
	Type        string `json:"type" yaml:"type"`
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Placeholder string `json:"placeholder,omitempty" yaml:"placeholder,omitempty"`
	Default     any    `json:"default,omitempty" yaml:"default,omitempty"`

	Choices []string `json:"choices,omitempty" yaml:"choices,omitempty"`
	Label   string   `json:"label,omitempty" yaml:"label,omitempty"`
}

func (c Command) CheckMissingParams(with map[string]any) error {
	for _, param := range c.Params {
		if _, ok := with[param.Name]; !ok && param.Default == nil {
			return fmt.Errorf("missing param: %s", param.Name)
		}
	}

	return nil
}

type CmdPayload struct {
	Args  map[string]any
	Dir   string
	Query string
}

func (c Command) Cmd(payload CmdPayload) (*exec.Cmd, error) {
	if err := c.CheckMissingParams(payload.Args); err != nil {
		return nil, err
	}

	paramMap := make(map[string]any)
	for _, param := range c.Params {
		input, ok := payload.Args[param.Name]
		if !ok {
			return nil, fmt.Errorf("param %s was not provided", param.Name)
		}

		switch param.Type {
		case "string":
			value, ok := input.(string)
			if !ok {
				return nil, fmt.Errorf("%s type was not bool", param.Name)
			}

			value, err := syntax.Quote(value, syntax.LangPOSIX)
			if err != nil {
				return nil, err
			}
			paramMap[param.Name] = value
		case "file", "directory":
			value, ok := input.(string)
			if !ok {
				return nil, fmt.Errorf("%s type was not bool", param.Name)
			}

			if strings.HasPrefix(value, "~") {
				homedir, err := os.UserHomeDir()
				if err != nil {
					return nil, err
				}
				value = strings.Replace(value, "~", homedir, 1)
			}

			value, err := syntax.Quote(value, syntax.LangPOSIX)
			if err != nil {
				return nil, err
			}
			paramMap[param.Name] = value
		case "boolean":
			value, ok := input.(bool)
			if !ok {
				return nil, fmt.Errorf("%s type was not bool", param.Name)
			}

			paramMap[param.Name] = value
		}
	}

	rendered, err := utils.RenderString(c.Exec, template.FuncMap{
		"param": func(name string) any {
			return paramMap[name]
		},
		"query": func() string {
			return payload.Query
		},
	})
	if err != nil {
		return nil, err
	}

	fields, err := shell.Fields(rendered, nil)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Env = os.Environ()
	for _, param := range c.Params {
		if param.Env != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", param.Env, payload.Args[param.Name]))
		}
	}

	for key, value := range c.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	cmd.Dir = payload.Dir

	return cmd, nil
}

type Page struct {
	Type    string   `json:"type"`
	Title   string   `json:"title"`
	Actions []Action `json:"actions,omitempty"`

	*Detail
	*List
}

type Detail struct {
	Content Preview `json:"content,omitempty"`
}

type List struct {
	ShowPreview   bool       `json:"showPreview,omitempty" yaml:"showPreview"`
	GenerateItems bool       `json:"generateItems,omitempty" yaml:"generateItems"`
	Items         []ListItem `json:"items"`
	EmptyMessage  string     `json:"emptyMessage,omitempty" yaml:"emptyMessage"`
}

type Preview struct {
	Text     string         `json:"text,omitempty"`
	Language string         `json:"language,omitempty"`
	Command  string         `json:"command,omitempty"`
	With     map[string]any `json:"with,omitempty"`
}

type ListItem struct {
	Id          string   `json:"id"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Preview     *Preview `json:"preview,omitempty"`
	Accessories []string `json:"accessories,omitempty"`
	Actions     []Action `json:"actions"`
}

type Action struct {
	Title    string `json:"title"`
	Type     string `json:"type"`
	Shortcut string `json:"shortcut,omitempty"`

	Text string `json:"text,omitempty"`

	Url  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`

	Command   string         `json:"command,omitempty"`
	With      map[string]Arg `json:"with,omitempty"`
	OnSuccess string         `json:"onSuccess,omitempty"`
}
