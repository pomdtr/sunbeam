package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/sunbeamlauncher/sunbeam/utils"
	"gopkg.in/yaml.v3"
)

type Script struct {
	Name        string
	Command     string        `json:"command" yaml:"command"`
	Description string        `json:"description" yaml:"description"`
	Preferences []ScriptParam `json:"preferences" yaml:"preferences"`
	Mode        string        `json:"mode" yaml:"mode"`
	Cwd         string        `json:"cwd" yaml:"cwd"`
	Params      []ScriptParam `json:"params" yaml:"params"`
	ShowPreview bool          `json:"showPreview" yaml:"showPreview"`
}

func (s Script) IsPage() bool {
	return s.Mode == "filter" || s.Mode == "generator" || s.Mode == "detail"
}

type Optional[T any] struct {
	Defined bool
	Value   T
}

// UnmarshalJSON is implemented by deferring to the wrapped type (T).
// It will be called only if the value is defined in the JSON payload.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Defined = true
	return json.Unmarshal(data, &o.Value)
}

func (o *Optional[T]) UnmarshalYAML(value *yaml.Node) (err error) {
	o.Defined = true
	return value.Decode(&o.Value)
}

type ScriptParam struct {
	Name        string           `json:"name" yaml:"name"`
	Type        string           `json:"type"`
	Optional    Optional[bool]   `json:"optional"`
	Title       string           `json:"title"`
	Placeholder Optional[string] `json:"placeholder"`
	Default     Optional[any]    `json:"defaultValue"`

	TrueSubstitution  string `json:"trueSubstitution"`
	FalseSubstitution string `json:"falseSubstitution"`

	Data []struct {
		Title string `json:"title,omitempty"`
		Value string `json:"value,omitempty"`
	} `json:"data,omitempty"`
	Label string `json:"label"`
}

type ScriptInput struct {
	Value any
	ScriptParam
}

func (si ScriptInput) GetValue() (string, error) {
	if si.Value == nil && !si.Optional.Value {
		return "", fmt.Errorf("required value is empty")
	}

	switch si.Type {
	case "checkbox":
		var value bool
		if si.Value != nil {
			if v, ok := si.Value.(bool); ok {
				value = v
			} else {
				return "", fmt.Errorf("invalid value for checkbox")
			}
		} else if si.Default.Value != nil {
			if v, ok := si.Default.Value.(bool); ok {
				value = v
			} else {
				return "", fmt.Errorf("invalid value for checkbox")
			}
		}

		if value {
			return si.TrueSubstitution, nil
		}
		return si.FalseSubstitution, nil
	default:
		var value string
		if si.Value != nil {
			if v, ok := si.Value.(string); ok {
				value = v
			} else {
				return "", fmt.Errorf("invalid value for checkbox")
			}
		} else if si.Default.Value != nil {
			if v, ok := si.Default.Value.(string); ok {
				value = v
			} else {
				return "", fmt.Errorf("invalid value for checkbox")
			}
		}

		if si.Type == "file" || si.Type == "directory" {
			if v, err := utils.ResolvePath(value); err != nil {
				return "", err
			} else {
				value = v
			}
		}

		return shellescape.Quote(value), nil
	}
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

func (s Script) Cmd(with map[string]string) (string, error) {
	var err error

	funcMap := template.FuncMap{}

	for sanitizedKey, value := range with {
		value := value
		sanitizedKey = strings.Replace(sanitizedKey, "-", "_", -1)
		funcMap[sanitizedKey] = func() string {
			return value
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

	Target string `json:"target,omitempty" yaml:"target"`

	Extension string `json:"extension,omitempty" yaml:"extension"`
	Script    string `json:"script,omitempty" yaml:"script"`
	Command   string `json:"command,omitempty" yaml:"command"`
	Dir       string

	OnSuccess string                 `json:"onSuccess,omitempty" yaml:"onSuccess"`
	With      map[string]ScriptInput `json:"with,omitempty" yaml:"with"`
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
