package types

import (
	_ "embed"
)

//go:embed typescript/index.d.ts
var TypeScript string

type PageType string

const (
	DetailPage PageType = "detail"
	ListPage   PageType = "list"
	FormPage   PageType = "form"
)

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

type PreviewType string

const (
	StaticPreviewType PreviewType = "static"
	DynamicPageType   PreviewType = "dynamic"
)

type Preview struct {
	Type     PreviewType `json:"type,omitempty" yaml:"type,omitempty"`
	Language string      `json:"language,omitempty" yaml:"language,omitempty"`
	Text     string      `json:"text,omitempty" yaml:"text,omitempty"`
	Command  string      `json:"command,omitempty" yaml:"command,omitempty"`
	Dir      string      `json:"dir,omitempty" yaml:"dir,omitempty"`
}

type ListItem struct {
	Id          string   `json:"id,omitempty" yaml:"id,omitempty"`
	Title       string   `json:"title" yaml:"title"`
	Subtitle    string   `json:"subtitle,omitempty" yaml:"subtitle,omitempty"`
	Preview     *Preview `json:"preview,omitempty" yaml:"preview,omitempty"`
	Accessories []string `json:"accessories,omitempty" yaml:"accessories,omitempty"`
	Actions     []Action `json:"actions,omitempty" yaml:"actions,omitempty"`
}

type FormInputType string

const (
	TextFieldInput FormInputType = "textfield"
	TextAreaInput  FormInputType = "textarea"
	DropDownInput  FormInputType = "dropdown"
	CheckboxInput  FormInputType = "checkbox"
)

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

type ActionType string

const (
	CopyAction     = "copy-text"
	OpenPathAction = "open-path"
	OpenUrlAction  = "open-url"
	PushPageAction = "push-page"
	RunAction      = "run-command"
	ReloadAction   = "reload-page"
)

type OnSuccessType string

const (
	ReloadOnSuccess OnSuccessType = "reload"
	ExitOnSuccess   OnSuccessType = "exit"
)

type TargetType string

const (
	StaticTarget  TargetType = "static"
	DynamicTarget TargetType = "dynamic"
)

type Target struct {
	Type    TargetType `json:"type" yaml:"type"`
	Path    string     `json:"path,omitempty" yaml:"path,omitempty"`
	Input   string     `json:"input,omitempty" yaml:"input,omitempty"`
	Command string     `json:"command,omitempty" yaml:"command,omitempty"`
	Dir     string     `json:"dir,omitempty" yaml:"dir,omitempty"`
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
	Page *Target `json:"page,omitempty" yaml:"page,omitempty"`

	// open
	Url string `json:"url,omitempty" yaml:"url,omitempty"`

	// run
	Command   string        `json:"command,omitempty" yaml:"command,omitempty"`
	Input     string        `json:"input,omitempty" yaml:"input,omitempty"`
	Dir       string        `json:"dir,omitempty" yaml:"dir,omitempty"`
	OnSuccess OnSuccessType `json:"onSuccess,omitempty" yaml:"onSuccess,omitempty"`
}
