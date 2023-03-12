package schemas

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"

	_ "embed"

	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
)

//go:embed schema.json
var schemaString string

var Schema *jsonschema.Schema

func init() {
	var err error

	compiler := jsonschema.NewCompiler()

	if err = compiler.AddResource("https://pomdtr.github.io/sunbeam/schemas/page.json", strings.NewReader(schemaString)); err != nil {
		panic(err)
	}
	Schema, err = compiler.Compile("https://pomdtr.github.io/sunbeam/schemas/page.json")
	if err != nil {
		panic(err)
	}
}

func Validate(bytes []byte) error {
	var v any
	if err := json.Unmarshal(bytes, &v); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	return Schema.Validate(v)
}

type PageType int

const (
	Unknown PageType = iota
	DetailPage
	ListPage
)

func (p *PageType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case "detail":
		*p = DetailPage
	case "list":
		*p = ListPage
	default:
		return fmt.Errorf("unknown page type: %s", s)
	}

	return nil
}

type Response struct {
	Type    PageType `json:"type"`
	Title   string   `json:"title"`
	Actions []Action `json:"actions"`

	*Detail
	*List
}

type Detail struct {
	Content Preview `json:"content"`
}

type List struct {
	ShowPreview   bool       `json:"showPreview"`
	GenerateItems bool       `json:"generateItems"`
	Items         []ListItem `json:"items"`
}

type Preview struct {
	Text     string   `json:"text"`
	Language string   `json:"language"`
	Command  string   `json:"command"`
	Args     []string `json:"args"`
}

type ListItem struct {
	Id          string   `json:"id"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Preview     *Preview `json:"preview"`
	Accessories []string `json:"accessories"`
	Actions     []Action `json:"actions"`
}

type FormInputType int

const (
	UnknownFormInput FormInputType = iota
	TextField
	TextArea
	DropDown
)

func (input *FormInputType) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return fmt.Errorf("unable to unmarshal form input type: %w", err)
	}

	switch s {
	case "text":
		*input = TextField
	case "textarea":
		*input = TextArea
	case "dropdown":
		*input = DropDown
	default:
		return fmt.Errorf("unknown form input type: %s", s)
	}

	return nil
}

type FormInput struct {
	Name        string        `json:"name"`
	Type        FormInputType `json:"type"`
	Title       string        `json:"title"`
	Placeholder string        `json:"placeholder"`
	Default     string        `json:"default"`

	Choices []string `json:"choices"`
	Label   string   `json:"label"`
}

type ActionType int

const (
	UnknownAction ActionType = iota
	CopyAction
	OpenAction
	ReadAction
	RunAction
	ReloadAction
)

func (a *ActionType) UnmarshalJSON(bytes []byte) error {
	var rawType string
	if err := json.Unmarshal(bytes, &rawType); err != nil {
		return err
	}

	switch rawType {
	case "copy":
		*a = CopyAction
	case "open":
		*a = OpenAction
	case "read":
		*a = ReadAction
	case "run":
		*a = RunAction
	case "reload":
		*a = ReloadAction
	default:
		return fmt.Errorf("unknown action type: %s", rawType)
	}

	return nil
}

type OnSuccessType int

const (
	ExitOnSuccess OnSuccessType = iota
	PushOnSuccess
	ReloadOnSuccess
)

func (o *OnSuccessType) UnmarshalJSON(bytes []byte) error {
	var rawType string
	if err := json.Unmarshal(bytes, &rawType); err != nil {
		return err
	}

	switch rawType {
	case "push":
		*o = PushOnSuccess
	case "reload":
		*o = ReloadOnSuccess
	default:
		return fmt.Errorf("unknown on success type: %s", rawType)
	}

	return nil
}

type Action struct {
	RawTitle string     `json:"title"`
	Shortcut string     `json:"shortcut"`
	Type     ActionType `json:"type"`

	Inputs []FormInput `json:"inputs"`

	// edit
	Page string `json:"page"`

	// open
	Target string `json:"target"`

	// copy
	Text string `json:"text"`

	// run / push
	Command string   `json:"command"`
	Args    []string `json:"args"`

	// run
	OnSuccess OnSuccessType `json:"onSuccess"`
}

func (a Action) Title() string {
	if a.RawTitle != "" {
		return a.RawTitle
	}
	switch a.Type {
	case CopyAction:
		return "Copy"
	case OpenAction:
		return "Open"
	case ReadAction:
		return "Read"
	case RunAction:
		return "Run"
	case ReloadAction:
		return "Reload"
	default:
		return "Unknown"
	}
}
