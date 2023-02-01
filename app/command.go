package app

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
	"github.com/pomdtr/sunbeam/utils"
	"gopkg.in/yaml.v3"
)

type Command struct {
	Name        string  `json:"name"`
	Exec        string  `json:"exec,omitempty" yaml:"exec,omitempty"`
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
	Env   map[string]string
	With  map[string]any
	Stdin string
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
	cmd.Stdin = strings.NewReader(payload.Stdin)

	cmd.Env = append(cmd.Env, os.Environ()...)
	for key, env := range payload.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, env))
	}

	return cmd, nil
}

func (c Command) Run(payload CommandPayload, root *url.URL) ([]byte, error) {
	var err error

	if root.Scheme != "file" {
		payload, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal command params: %v", err)
		}

		url := url.URL{
			Scheme: root.Scheme,
			Host:   root.Host,
			Path:   path.Join(root.Path, "commands", c.Name),
		}

		res, err := http.Post(url.String(), "application/json", bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("failed to execute command: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to execute command: %s", res.Status)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read command output: %v", err)
		}

		return body, nil
	}

	cmd, err := c.Cmd(payload, root.Path)
	if err != nil {
		return nil, err
	}

	output, err := cmd.Output()
	if err != nil {
		err, ok := err.(*exec.ExitError)
		if !ok {
			return nil, err
		}
		return nil, fmt.Errorf("%s: %s", err, err.Stderr)
	}

	return output, nil
}

type Page struct {
	Type  string `json:"type"`
	Title string `json:"title"`

	Detail
	List
}

type Detail struct {
	Content Preview  `json:"content"`
	Actions []Action `json:"actions,omitempty"`
}

type List struct {
	ShowPreview   bool       `json:"showPreview,omitempty" yaml:"showPreview"`
	GenerateItems bool       `json:"generateItems,omitempty" yaml:"generateItems"`
	Items         []ListItem `json:"items"`
}

type Preview struct {
	Text     string         `json:"text"`
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

	Command   string                  `json:"command,omitempty"`
	With      map[string]CommandInput `json:"with,omitempty"`
	OnSuccess string                  `json:"onSuccess,omitempty"`
}
