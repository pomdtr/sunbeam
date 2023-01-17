package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

type Page struct {
	Type        string
	ShowPreview bool `json:"showPreview" yaml:"showPreview"`
	IsGenerator bool `json:"isGenerator" yaml:"isGenerator"`
}

type Command struct {
	Name        string
	Exec        string
	Description string
	Preferences []ScriptInput
	Inputs      []ScriptInput
	Page        Page

	OnSuccess string `json:"onSuccess" yaml:"onSuccess"`
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

type ScriptInput struct {
	Name        string
	Type        string
	Title       string
	Placeholder Optional[string]
	Default     Optional[any]

	Data []struct {
		Title string
		Value string
	}
	Label string
}

type ScriptInputWithValue struct {
	Value any
	ScriptInput
}

func (si ScriptInputWithValue) GetValue() (any, error) {
	if si.Value == nil {
		return "", fmt.Errorf("required value %s is empty", si.Name)
	}

	var value string
	if si.Value != nil {
		if v, ok := si.Value.(string); ok {
			value = v
		} else {
			return "", fmt.Errorf("invalid value")
		}
	} else if si.Default.Value != nil {
		if v, ok := si.Default.Value.(string); ok {
			value = v
		} else {
			return "", fmt.Errorf("invalid value")
		}
	}

	if si.Type == "file" || si.Type == "directory" {
		if v, err := utils.ResolvePath(value); err == nil {
			return shellescape.Quote(v), nil
		} else {
			return "", err
		}
	}

	return shellescape.Quote(value), nil
}

func (si *ScriptInputWithValue) UnmarshalYAML(value *yaml.Node) (err error) {
	err = value.Decode(&si.ScriptInput)
	if err == nil {
		return
	}

	return value.Decode(&si.Value)
}

func (si *ScriptInputWithValue) UnmarshalJSON(bytes []byte) (err error) {
	err = json.Unmarshal(bytes, &si.ScriptInput)
	if err == nil {
		return
	}

	return json.Unmarshal(bytes, &si.Value)
}

func (s Command) Cmd(with map[string]any) (string, error) {
	var err error

	funcMap := template.FuncMap{}

	for sanitizedKey, value := range with {
		value := value
		sanitizedKey = strings.Replace(sanitizedKey, "-", "_", -1)
		funcMap[sanitizedKey] = func() any {
			return value
		}
	}

	rendered, err := utils.RenderString(s.Exec, funcMap)
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

	Extension string
	Command   string

	OnSuccess string `json:"onSuccess"`
	With      map[string]ScriptInputWithValue
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
