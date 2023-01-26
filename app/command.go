package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
	"github.com/pomdtr/sunbeam/utils"
	"gopkg.in/yaml.v3"
)

type Command struct {
	Exec        string  `json:"exec,omitempty" yaml:"exec,omitempty"`
	Interactive bool    `json:"interactive,omitempty" yaml:"interactive,omitempty"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Params      []Param `json:"params,omitempty" yaml:"params,omitempty"`
	OnSuccess   string  `json:"onSuccess,omitempty" yaml:"onSuccess,omitempty"`
}

type CommandInput struct {
	Value    any
	FormItem FormItem
}

func (i *CommandInput) UnmarshalJSON(b []byte) error {
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
		i.FormItem = input
		return nil
	}

	return fmt.Errorf("invalid input: %s", b)
}

func (i *CommandInput) UnmarshalYAML(node *yaml.Node) error {
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
		i.FormItem = input
		return nil
	}

	return fmt.Errorf("invalid input: %s", node.Value)
}

func (i CommandInput) MarshalYAML() (interface{}, error) {
	if i.Value != nil {
		return i.Value, nil
	}
	return i.FormItem, nil
}

type Param struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Default     any      `json:"default,omitempty" yaml:"default,omitempty"`
	Description string   `json:",omitempty" yaml:",omitempty"`
	Pattern     string   `json:",omitempty" yaml:",omitempty"`
	Enum        []string `json:",omitempty" yaml:",omitempty"`
}

type FormItem struct {
	Type        string
	Title       string `json:",omitempty" yaml:",omitempty"`
	Placeholder string `json:",omitempty" yaml:",omitempty"`
	Default     any    `json:",omitempty" yaml:",omitempty"`

	Choices []string `json:",omitempty" yaml:",omitempty"`
	Label   string   `json:"label,omitempty" yaml:"label,omitempty"`
}

type CommandPayload struct {
	Input string
	Env   map[string]string
	With  map[string]any
}

func (c Command) CheckMissingParams(with map[string]any) error {
	for _, param := range c.Params {
		if _, ok := with[param.Name]; !ok && param.Default == nil {
			return fmt.Errorf("missing param: %s", param.Name)
		}
	}

	return nil
}

func (c Command) Cmd(payload CommandPayload, dir string) (*exec.Cmd, error) {
	var err error

	funcMap := template.FuncMap{}
	for _, spec := range c.Params {
		input, ok := payload.With[spec.Name]
		if !ok {
			return nil, fmt.Errorf("param %s was not provided", spec.Name)
		}

		sanitizedKey := strings.Replace(spec.Name, "-", "_", -1)
		switch spec.Type {
		case "string":
			value, ok := input.(string)
			if !ok {
				return nil, fmt.Errorf("%s type was not bool", spec.Name)
			}

			funcMap[sanitizedKey] = func() string { return shellescape.Quote(value) }
		case "file", "directory":
			value, ok := input.(string)
			if !ok {
				return nil, fmt.Errorf("%s type was not bool", spec.Name)
			}

			if strings.HasPrefix(value, "~") {
				homedir, err := os.UserHomeDir()
				if err != nil {
					return nil, err
				}
				value = strings.Replace(value, "~", homedir, 1)
			}

			funcMap[sanitizedKey] = func() string { return shellescape.Quote(value) }
		case "boolean":
			value, ok := input.(bool)
			if !ok {
				return nil, fmt.Errorf("%s type was not bool", spec.Name)
			}

			funcMap[sanitizedKey] = func() bool { return value }
		}
	}

	rendered, err := utils.RenderString(c.Exec, funcMap)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("sh", "-c", rendered)
	cmd.Dir = dir

	cmd.Env = append(cmd.Env, os.Environ()...)
	for key, env := range payload.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, env))
	}

	cmd.Stdin = strings.NewReader(payload.Input)

	return cmd, nil
}

type Page struct {
	Type  string `json:"type"`
	Title string `json:"title"`

	Detail
	List
}

type Detail struct {
	Preview *Preview `json:"preview,omitempty"`
	Actions []Action `json:"actions,omitempty"`
}

type List struct {
	ShowPreview bool       `json:"showPreview,omitempty" yaml:"showPreview"`
	Items       []ListItem `json:"items"`
}

type Preview struct {
	Text string `json:"text"`
	PreviewCommand
}

type PreviewCommand struct {
	Command string         `json:"command"`
	With    map[string]any `json:"with"`
}

func (p *Preview) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		p.Text = s
		return nil
	}

	var cmd PreviewCommand
	if err := json.Unmarshal(b, &cmd); err == nil {
		p.PreviewCommand = cmd
		return nil
	}

	return fmt.Errorf("invalid preview: %s", b)
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

	Command   string                  `json:"command,omitempty"`
	With      map[string]CommandInput `json:"with,omitempty"`
	OnSuccess string                  `json:"onSuccess,omitempty"`
}
