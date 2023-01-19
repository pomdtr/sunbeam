package app

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/pomdtr/sunbeam/utils"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type Page struct {
	Type        string `json:"type"`
	ShowPreview bool   `json:"showPreview,omitempty" yaml:"showPreview"`
	IsGenerator bool   `json:"isGenerator,omitempty" yaml:"isGenerator"`
}

type Command struct {
	Exec        string      `json:"exec,omitempty"`
	Url         string      `json:"url"`
	Description string      `json:"description,omitempty"`
	Preferences []FormInput `json:"preferences,omitempty"`
	Inputs      []FormInput `json:"inputs,omitempty"`
	Page        *Page       `json:"page,omitempty"`

	OnSuccess string `json:"onSuccess,omitempty" yaml:"onSuccess"`
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

type CommandInput struct {
	Dir   string
	Stdin string
	Env   []string
	With  map[string]any
}

func (c Command) Cmd(input CommandInput) (*exec.Cmd, error) {
	var err error

	funcMap := template.FuncMap{}

	for sanitizedKey, value := range input.With {
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
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, input.Env...)

	cmd.Dir = input.Dir

	if input.Stdin != "" {
		cmd.Stdin = strings.NewReader(input.Stdin)
	}

	return cmd, nil
}

type Detail struct {
	Actions []ScriptAction `json:"actions"`
	DetailData
}

type DetailData struct {
	Preview   string           `json:"preview"`
	Metadatas []ScriptMetadata `json:"metadatas"`
}

type ScriptMetadata struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type ScriptItem struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	DetailData
	Accessories []string       `json:"accessories"`
	Actions     []ScriptAction `json:"actions"`
}

func (li ScriptItem) PreviewCommand() *exec.Cmd {
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

	OnSuccess string `json:"onSuccess"`
	With      map[string]any
}

//go:embed schemas/listitem.json
var itemSchemaString string
var itemSchema *jsonschema.Schema

func init() {
	compiler := jsonschema.NewCompiler()
	compiler.AddResource("listitem.json", strings.NewReader(itemSchemaString))
	itemSchema = compiler.MustCompile("listitem.json")
}

func ParseListItems(output string) (items []ScriptItem, err error) {
	rows := strings.Split(output, "\n")
	for _, row := range rows {
		if row == "" {
			continue
		}
		var data any
		err = json.Unmarshal([]byte(row), &data)
		if err != nil {
			return
		}
		err = itemSchema.Validate(data)
		if err != nil {
			return
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
