package types

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

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

func (p PageType) MarshalJSON() ([]byte, error) {
	switch p {
	case DetailPage:
		return json.Marshal("detail")
	case ListPage:
		return json.Marshal("list")
	default:
		return nil, fmt.Errorf("unknown page type: %d", p)
	}
}

func (p *PageType) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
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

func (p PageType) MarshalYAML() (interface{}, error) {
	switch p {
	case DetailPage:
		return "detail", nil
	case ListPage:
		return "list", nil
	default:
		return nil, fmt.Errorf("unknown page type: %d", p)
	}
}

type Page struct {
	Type    PageType `json:"type" yaml:"type"`
	Title   string   `json:"title,omitempty" yaml:"title,omitempty"`
	Actions []Action `json:"actions,omitempty" yaml:"actions,omitempty"`

	// Detail page
	Text     string `json:"text,omitempty" yaml:"text,omitempty"`
	Command  string `json:"command,omitempty" yaml:"command,omitempty"`
	Language string `json:"language,omitempty" yaml:"language,omitempty"`

	// List page
	ShowDetail    bool       `json:"showDetail,omitempty" yaml:"showDetail,omitempty"`
	GenerateItems bool       `json:"generateItems,omitempty" yaml:"generateItems,omitempty"`
	EmptyText     string     `json:"emptyText,omitempty" yaml:"emptyText,omitempty"`
	Items         []ListItem `json:"items,omitempty" yaml:"items,omitempty"`
}

type ListItem struct {
	Id          string   `json:"id,omitempty" yaml:"id,omitempty"`
	Alias       string   `json:"alias,omitempty" yaml:"alias,omitempty"`
	Title       string   `json:"title" yaml:"title"`
	Subtitle    string   `json:"subtitle,omitempty" yaml:"subtitle,omitempty"`
	Detail      *Detail  `json:"detail,omitempty" yaml:"detail,omitempty"`
	Accessories []string `json:"accessories,omitempty" yaml:"accessories,omitempty"`
	Actions     []Action `json:"actions,omitempty" yaml:"actions,omitempty"`
}

type Detail struct {
	Text     string `json:"text" yaml:"text"`
	Command  string `json:"command" yaml:"command"`
	Language string `json:"language" yaml:"language"`
}

type FormInputType int

const (
	UnknownFormInput FormInputType = iota
	TextFieldInput
	TextAreaInput
	DropDownInput
)

func (input *FormInputType) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return fmt.Errorf("unable to unmarshal form input type: %w", err)
	}

	switch s {
	case "textfield":
		*input = TextFieldInput
	case "textarea":
		*input = TextAreaInput
	case "dropdown":
		*input = DropDownInput
	default:
		return fmt.Errorf("unknown form input type: %s", s)
	}

	return nil
}

func (input *FormInputType) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return fmt.Errorf("unable to unmarshal form input type: %w", err)
	}

	switch s {
	case "textfield":
		*input = TextFieldInput
	case "textarea":
		*input = TextAreaInput
	case "dropdown":
		*input = DropDownInput
	default:
		return fmt.Errorf("unknown form input type: %s", s)
	}

	return nil
}

func (input FormInputType) MarshalJSON() ([]byte, error) {
	switch input {
	case TextFieldInput:
		return json.Marshal("textfield")
	case TextAreaInput:
		return json.Marshal("textarea")
	case DropDownInput:
		return json.Marshal("dropdown")
	default:
		return nil, fmt.Errorf("unknown form input type: %d", input)
	}
}

func (input FormInputType) MarshalYAML() (interface{}, error) {
	switch input {
	case TextFieldInput:
		return "textfield", nil
	case TextAreaInput:
		return "textarea", nil
	case DropDownInput:
		return "dropdown", nil
	default:
		return nil, fmt.Errorf("unknown form input type: %d", input)
	}
}

type Input struct {
	Name        string        `json:"name" yaml:"name"`
	Type        FormInputType `json:"type" yaml:"type"`
	Title       string        `json:"title" yaml:"title"`
	Placeholder string        `json:"placeholder,omitempty" yaml:"placeholder,omitempty"`
	Default     string        `json:"default,omitempty" yaml:"default,omitempty"`

	// Only for dropdown
	Choices []string `json:"choices,omitempty" yaml:"choices,omitempty"`
}

type ActionType int

const (
	UnknownAction ActionType = iota
	CopyAction
	OpenAction
	ReadAction
	EditAction
	RunAction
	HttpAction
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
	case "http":
		*a = HttpAction
	case "edit":
		*a = EditAction
	case "reload":
		*a = ReloadAction
	default:
		return fmt.Errorf("unknown action type: %s", rawType)
	}

	return nil
}

func (a ActionType) MarshalJSON() ([]byte, error) {
	switch a {
	case CopyAction:
		return json.Marshal("copy")
	case OpenAction:
		return json.Marshal("open")
	case ReadAction:
		return json.Marshal("read")
	case EditAction:
		return json.Marshal("edit")
	case RunAction:
		return json.Marshal("run")
	case HttpAction:
		return json.Marshal("http")
	case ReloadAction:
		return json.Marshal("reload")
	default:
		return nil, fmt.Errorf("unknown action type: %d", a)
	}
}

func (a *ActionType) UnmarshalYAML(node *yaml.Node) error {
	var rawType string
	if err := node.Decode(&rawType); err != nil {
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
	case "http":
		*a = HttpAction
	case "edit":
		*a = EditAction
	case "reload":
		*a = ReloadAction
	default:
		return fmt.Errorf("unknown action type: %s", rawType)
	}

	return nil
}

func (a ActionType) MarshalYAML() (interface{}, error) {
	switch a {
	case CopyAction:
		return "copy", nil
	case OpenAction:
		return "open", nil
	case ReadAction:
		return "read", nil
	case EditAction:
		return "edit", nil
	case RunAction:
		return "run", nil
	case HttpAction:
		return "http", nil
	case ReloadAction:
		return "reload", nil
	default:
		return nil, fmt.Errorf("unknown action type: %d", a)
	}
}

type OnSuccessType int

const (
	ExitOnSuccess OnSuccessType = iota
	PushOnSuccess
	ReplaceOnSuccess
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
	case "replace":
		*o = ReplaceOnSuccess
	default:
		return fmt.Errorf("unknown on success type: %s", rawType)
	}

	return nil
}

func (o OnSuccessType) MarshalJSON() ([]byte, error) {
	switch o {
	case PushOnSuccess:
		return json.Marshal("push")
	case ReloadOnSuccess:
		return json.Marshal("reload")
	case ReplaceOnSuccess:
		return json.Marshal("replace")
	default:
		return nil, fmt.Errorf("unknown on success type: %d", o)
	}
}

func (o *OnSuccessType) UnmarshalYAML(node *yaml.Node) error {
	var rawType string
	if err := node.Decode(&rawType); err != nil {
		return err
	}

	switch rawType {
	case "push":
		*o = PushOnSuccess
	case "reload":
		*o = ReloadOnSuccess
	case "replace":
		*o = ReplaceOnSuccess
	default:
		return fmt.Errorf("unknown onSuccess type: %s", rawType)
	}

	return nil
}

func (o OnSuccessType) MarshalYAML() (interface{}, error) {
	switch o {
	case PushOnSuccess:
		return "push", nil
	case ReloadOnSuccess:
		return "reload", nil
	case ReplaceOnSuccess:
		return "replace", nil
	default:
		return nil, fmt.Errorf("unknown onSuccess type: %d", o)
	}
}

type Action struct {
	Title    string     `json:"title,omitempty" yaml:"title,omitempty"`
	Shortcut string     `json:"shortcut,omitempty" yaml:"shortcut,omitempty"`
	Type     ActionType `json:"type" yaml:"type"`

	// copy
	Text string `json:"text,omitempty" yaml:"text,omitempty"`

	// edit / open
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// open / http
	Url string `json:"url,omitempty" yaml:"url,omitempty"`

	// http
	Method  string            `json:"method,omitempty" yaml:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body    string            `json:"body,omitempty" yaml:"body,omitempty"`

	// run
	Command   string        `json:"command,omitempty" yaml:"command,omitempty"`
	Inputs    []Input       `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	OnSuccess OnSuccessType `json:"onSuccess,omitempty" yaml:"onSuccess,omitempty"`
}
