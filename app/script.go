package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

type Script struct {
	Command     string        `json:"command" yaml:"command"`
	Description string        `json:"description" yaml:"description"`
	Preferences []ScriptParam `json:"preferences" yaml:"preferences"`
	Mode        string        `json:"mode" yaml:"mode"`
	Cwd         string        `json:"cwd" yaml:"cwd"`
	Params      []ScriptParam `json:"params" yaml:"params"`
	Title       string        `json:"title" yaml:"title"`
	ShowPreview bool          `json:"showPreview" yaml:"showPreview"`
}

func (s Script) IsPage() bool {
	return s.Mode == "filter" || s.Mode == "generator" || s.Mode == "detail"
}

type ScriptParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Title       string `json:"title"`
	Placeholder string `json:"placeholder"`
	Default     any    `json:"default"`
	Data        []struct {
		Title string `json:"title,omitempty"`
		Value string `json:"value,omitempty"`
	} `json:"data,omitempty"`
	Label string `json:"label"`
}

type ScriptInput struct {
	Value any
	ScriptParam
}

func (si *ScriptInput) UnmarshalYAML(value *yaml.Node) (err error) {
	err = value.Decode(&si.ScriptParam)
	if err == nil {
		return
	}

	return value.Decode(&si.Value)
}

func (si *ScriptInput) UnmarshalJSON(bytes []byte) (err error) {
	err = json.Unmarshal(bytes, &si.ScriptParam)
	if err == nil {
		return
	}

	return json.Unmarshal(bytes, &si.Value)
}

type ScriptInputs map[string]ScriptInput

func (s Script) Cmd(with map[string]any) (string, error) {
	var err error

	funcMap := template.FuncMap{}

	for _, param := range s.Params {
		param := param
		value, ok := with[param.Name]
		if !ok {
			if param.Required {
				return "", fmt.Errorf("missing required param %s", param.Name)
			}
			value = param.Default
		}

		funcMap[param.Name] = func() (any, error) {
			switch param.Type {
			case "directory", "file":
				value, ok := value.(string)
				if !ok {
					return "", fmt.Errorf("expected string for param %s", param.Name)
				}

				if strings.HasPrefix(value, "~") {
					homeDir, err := os.UserHomeDir()
					if err != nil {
						return nil, err
					}
					value = strings.Replace(value, "~", homeDir, 1)
				} else if value == "." {
					value, err = os.Getwd()
					if err != nil {
						return nil, err
					}
				} else if !filepath.IsAbs(value) {
					cwd, err := os.Getwd()
					if err != nil {
						return nil, err
					}
					value = filepath.Join(cwd, value)
				}

				return shellescape.Quote(value), nil
			case "checkbox":
				value, ok := value.(bool)
				if !ok {
					return nil, fmt.Errorf("expected boolean for param %s", param.Name)
				}
				return value, nil
			default:
				value, ok := value.(string)
				if !ok {
					return "", fmt.Errorf("expected string for param %s", param.Name)
				}

				return shellescape.Quote(value), nil
			}
		}
	}

	rendered, err := utils.RenderString(s.Command, funcMap)
	if err != nil {
		return "", err
	}
	return rendered, nil
}

type Detail struct {
	Actions []ScriptAction `json:"actions"`
	DetailData
}

type DetailData struct {
	Preview    string           `json:"preview"`
	PreviewCmd string           `json:"previewCmd"`
	Metadatas  []ScriptMetadata `json:"metadatas"`
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

	Silent    bool         `json:"silent,omitempty" yaml:"silent"`
	OnSuccess string       `json:"onSuccess,omitempty" yaml:"onSuccess"`
	With      ScriptInputs `json:"with,omitempty" yaml:"with"`
}

//go:embed listitem.json
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
