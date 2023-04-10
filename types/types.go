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
	FormPage
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
	case "form":
		*p = FormPage
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
	case FormPage:
		return json.Marshal("form")
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
	case "form":
		*p = FormPage
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
	case FormPage:
		return "form", nil
	default:
		return nil, fmt.Errorf("unknown page type: %d", p)
	}
}

type Page struct {
	Type    PageType `json:"type" yaml:"type"`
	Title   string   `json:"title,omitempty" yaml:"title,omitempty"`
	Actions []Action `json:"actions,omitempty" yaml:"actions,omitempty"`

	// Form page
	SubmitAction *Action `json:"submitAction,omitempty" yaml:"submitAction,omitempty"`

	// Detail page
	Preview *Preview `json:"preview,omitempty" yaml:"preview,omitempty"`

	// List page
	ShowPreview bool `json:"showPreview,omitempty" yaml:"showPreview,omitempty"`
	EmptyView   *struct {
		Text    string   `json:"text,omitempty" yaml:"text,omitempty"`
		Actions []Action `json:"actions,omitempty" yaml:"actions,omitempty"`
	} `json:"emptyView,omitempty" yaml:"emptyView,omitempty"`
	Items []ListItem `json:"items,omitempty" yaml:"items,omitempty"`
}

type Preview struct {
	Type     string `json:"type,omitempty" yaml:"type,omitempty"`
	Language string `json:"language,omitempty" yaml:"language,omitempty"`
	Text     string `json:"text,omitempty" yaml:"text,omitempty"`
	Command  string `json:"command,omitempty" yaml:"command,omitempty"`
	Dir      string `json:"dir,omitempty" yaml:"dir,omitempty"`
}

type ListItem struct {
	Id          string   `json:"id,omitempty" yaml:"id,omitempty"`
	Title       string   `json:"title" yaml:"title"`
	Subtitle    string   `json:"subtitle,omitempty" yaml:"subtitle,omitempty"`
	Preview     *Preview `json:"preview,omitempty" yaml:"preview,omitempty"`
	Accessories []string `json:"accessories,omitempty" yaml:"accessories,omitempty"`
	Actions     []Action `json:"actions,omitempty" yaml:"actions,omitempty"`
}

type FormInputType int

const (
	UnknownFormInput FormInputType = iota
	TextFieldInput
	TextAreaInput
	DropDownInput
	CheckboxInput
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
	case "checkbox":
		*input = CheckboxInput
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
	case "checkbox":
		*input = CheckboxInput
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
	case CheckboxInput:
		return json.Marshal("checkbox")
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
	case CheckboxInput:
		return "checkbox", nil
	default:
		return nil, fmt.Errorf("unknown form input type: %d", input)
	}
}

type DropDownItem struct {
	Title string `json:"title" yaml:"title"`
	Value string `json:"value" yaml:"value"`
}

type Input struct {
	Name        string        `json:"name" yaml:"name"`
	Type        FormInputType `json:"type" yaml:"type"`
	Title       string        `json:"title" yaml:"title"`
	Placeholder string        `json:"placeholder,omitempty" yaml:"placeholder,omitempty"`
	Default     any           `json:"default,omitempty" yaml:"default,omitempty"`

	// Only for dropdown
	Items []DropDownItem `json:"items,omitempty" yaml:"items,omitempty"`

	// Only for checkbox
	Label             string `json:"label,omitempty" yaml:"label,omitempty"`
	TrueSubstitution  string `json:"trueSubstitution,omitempty" yaml:"trueSubstitution,omitempty"`
	FalseSubstitution string `json:"falseSubstitution,omitempty" yaml:"falseSubstitution,omitempty"`
}

type ActionType int

const (
	UnknownAction ActionType = iota
	CopyAction
	OpenFileAction
	OpenUrlAction
	PushPageAction
	RunAction
	ReloadAction
)

func (a *ActionType) UnmarshalJSON(bytes []byte) error {
	var rawType string
	if err := json.Unmarshal(bytes, &rawType); err != nil {
		return err
	}

	switch rawType {
	case "copy-text":
		*a = CopyAction
	case "open-file":
		*a = OpenFileAction
	case "open-url":
		*a = OpenUrlAction
	case "push-page":
		*a = PushPageAction
	case "run-command":
		*a = RunAction
	case "reload-page":
		*a = ReloadAction
	default:
		return fmt.Errorf("unknown action type: %s", rawType)
	}

	return nil
}

func (a ActionType) MarshalJSON() ([]byte, error) {
	switch a {
	case CopyAction:
		return json.Marshal("copy-text")
	case OpenFileAction:
		return json.Marshal("open-file")
	case OpenUrlAction:
		return json.Marshal("open-url")
	case PushPageAction:
		return json.Marshal("push-page")
	case RunAction:
		return json.Marshal("run-command")
	case ReloadAction:
		return json.Marshal("reload-page")
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
	case "copy-text":
		*a = CopyAction
	case "open-file":
		*a = OpenFileAction
	case "open-url":
		*a = OpenUrlAction
	case "push-page":
		*a = PushPageAction
	case "run-command":
		*a = RunAction
	case "reload-page":
		*a = ReloadAction
	default:
		return fmt.Errorf("unknown action type: %s", rawType)
	}

	return nil
}

func (a ActionType) MarshalYAML() (interface{}, error) {
	switch a {
	case CopyAction:
		return "copy-text", nil
	case OpenFileAction:
		return "open-file", nil
	case OpenUrlAction:
		return "open-url", nil
	case PushPageAction:
		return "push-page", nil
	case RunAction:
		return "run-command", nil
	case ReloadAction:
		return "reload-page", nil
	default:
		return nil, fmt.Errorf("unknown action type: %d", a)
	}
}

type OnSuccessType int

const (
	UnknowOnSuccess OnSuccessType = iota
	PushOnSuccess
	ReloadOnSuccess
	ExitOnSuccess
)

func (o *OnSuccessType) UnmarshalJSON(bytes []byte) error {
	var rawType string
	if err := json.Unmarshal(bytes, &rawType); err != nil {
		return err
	}

	switch rawType {
	case "push":
		*o = PushOnSuccess
	case "exit":
		*o = ExitOnSuccess
	case "reload":
		*o = ReloadOnSuccess
	default:
		return fmt.Errorf("unknown on success type: %s", rawType)
	}

	return nil
}

func (o OnSuccessType) MarshalJSON() ([]byte, error) {
	switch o {
	case PushOnSuccess:
		return json.Marshal("push")
	case ExitOnSuccess:
		return json.Marshal("exit")
	case ReloadOnSuccess:
		return json.Marshal("reload")
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
	case "exit":
		*o = ExitOnSuccess
	case "reload":
		*o = ReloadOnSuccess
	default:
		return fmt.Errorf("unknown onSuccess type: %s", rawType)
	}

	return nil
}

func (o OnSuccessType) MarshalYAML() (interface{}, error) {
	switch o {
	case PushOnSuccess:
		return "push", nil
	case ExitOnSuccess:
		return "exit", nil
	case ReloadOnSuccess:
		return "replace", nil
	default:
		return nil, fmt.Errorf("unknown onSuccess type: %d", o)
	}
}

type PageRef struct {
	Type    string `json:"type" yaml:"type"`
	Path    string `json:"path" yaml:"path"`
	Command string `json:"command,omitempty" yaml:"command,omitempty"`
	Dir     string `json:"dir,omitempty" yaml:"dir,omitempty"`
}

type Action struct {
	Title  string     `json:"title,omitempty" yaml:"title,omitempty"`
	Type   ActionType `json:"type" yaml:"type"`
	Key    string     `json:"key,omitempty" yaml:"key,omitempty"`
	Inputs []Input    `json:"inputs,omitempty" yaml:"inputs,omitempty"`

	// copy
	Text string `json:"text,omitempty" yaml:"text,omitempty"`

	// open
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// push
	Page *PageRef `json:"page,omitempty" yaml:"page,omitempty"`

	// open
	Url string `json:"url,omitempty" yaml:"url,omitempty"`

	// run
	Command   string        `json:"command,omitempty" yaml:"command,omitempty"`
	Input     string        `json:"input,omitempty" yaml:"input,omitempty"`
	Dir       string        `json:"dir,omitempty" yaml:"dir,omitempty"`
	OnSuccess OnSuccessType `json:"onSuccess,omitempty" yaml:"onSuccess,omitempty"`
}
