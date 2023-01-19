package app

import (
	_ "embed"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/pomdtr/sunbeam/utils"
)

type Command struct {
	Exec        string      `json:"exec,omitempty"`
	Url         string      `json:"url"`
	Description string      `json:"description,omitempty"`
	Preferences []FormInput `json:"preferences,omitempty"`
	Inputs      []FormInput `json:"inputs,omitempty"`
	OnSuccess   string      `json:"onSuccess,omitempty" yaml:"onSuccess"`
}

type FormInput struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Placeholder string `json:"placeholder,omitempty"`
	Default     any    `json:"default,omitempty"`

	Choices []string
	Label   string `json:"label,omitempty"`
}

type CommandParams struct {
	Input string
	Env   []string
	With  map[string]any
}

func (c Command) Cmd(params CommandParams, dir string) (*exec.Cmd, error) {
	var err error

	funcMap := template.FuncMap{}

	for sanitizedKey, value := range params.With {
		value := value
		sanitizedKey = strings.Replace(sanitizedKey, "-", "_", -1)
		funcMap[sanitizedKey] = func() any {
			return value
		}
	}

	rendered, err := utils.RenderString(c.Exec, funcMap)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("sh", "-c", rendered)
	cmd.Dir = dir
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, params.Env...)

	if params.Input != "" {
		cmd.Stdin = strings.NewReader(params.Input)
	}

	return cmd, nil
}

type Detail struct {
	Preview   string     `json:"preview"`
	Metadatas []Metadata `json:"metadatas"`

	Actions []ScriptAction `json:"actions"`
}

type Metadata struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type Page struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Form
	List
	Detail
}

type Form struct {
	Inputs []FormInput `json:"inputs"`
	Target struct {
		Command string         `json:"command"`
		With    map[string]any `json:"with"`
	}
}

type List struct {
	ShowPreview bool       `json:"showPreview,omitempty" yaml:"showPreview"`
	Items       []ListItem `json:"items"`
}

type ListItem struct {
	Id          string         `json:"id"`
	Title       string         `json:"title"`
	Subtitle    string         `json:"subtitle"`
	Preview     string         `json:"preview"`
	Accessories []string       `json:"accessories"`
	Actions     []ScriptAction `json:"actions"`
}

func (li ListItem) PreviewCommand() *exec.Cmd {
	return nil
}

type ScriptAction struct {
	Title    string
	Type     string
	Shortcut string

	Text string

	Url  string
	Path string

	Command string
	With    map[string]any
}
