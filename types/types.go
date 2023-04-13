package types

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/shlex"
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

type Command struct {
	Name  string   `json:"name" yaml:"name"`
	Args  []string `json:"args,omitempty" yaml:"args,omitempty"`
	Input string   `json:"input,omitempty" yaml:"input,omitempty"`
	Dir   string   `json:"dir,omitempty" yaml:"dir,omitempty"`
}

func (c *Command) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		args, err := shlex.Split(s)
		if err != nil {
			return err
		}
		c.Name = args[0]
		c.Args = args[1:]
		return nil
	}

	var v any
	if err := json.Unmarshal(data, &v); err == nil {
		if v, ok := v.(map[string]interface{}); ok {
			if name, ok := v["name"].(string); ok {
				c.Name = name
			}
			if args, ok := v["args"].([]interface{}); ok {
				for _, arg := range args {
					c.Args = append(c.Args, arg.(string))
				}
			}
			if input, ok := v["input"].(string); ok {
				c.Input = input
			}
			if dir, ok := v["dir"].(string); ok {
				c.Dir = dir
			}
			return nil
		}

	}

	return fmt.Errorf("invalid command: %s", string(data))
}

func (c Command) cmd() *exec.Cmd {
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Dir = c.Dir
	cmd.Stdin = strings.NewReader(c.Input)

	return cmd
}

func (c Command) Run() error {
	cmd := c.cmd()
	return cmd.Run()
}

func (c Command) Output() ([]byte, error) {
	cmd := c.cmd()
	return cmd.Output()
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
	Command  *Command    `json:"command,omitempty" yaml:"command,omitempty"`
}

type ListItem struct {
	Id          string   `json:"id,omitempty" yaml:"id,omitempty"`
	Title       string   `json:"title" yaml:"title"`
	Subtitle    string   `json:"subtitle,omitempty" yaml:"subtitle,omitempty"`
	Data        any      `json:"data,omitempty" yaml:"data,omitempty"`
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

type TargetType string

const (
	StaticTarget  TargetType = "static"
	DynamicTarget TargetType = "dynamic"
)

type Target struct {
	Type    TargetType `json:"type" yaml:"type"`
	Path    string     `json:"path,omitempty" yaml:"path,omitempty"`
	Command *Command   `json:"command,omitempty" yaml:"command,omitempty"`
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
	Command         *Command `json:"command,omitempty" yaml:"command,omitempty"`
	ReloadOnSuccess bool     `json:"reloadOnSuccess,omitempty" yaml:"reloadOnSuccess,omitempty"`
}
