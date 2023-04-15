package types

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/shlex"
	"github.com/mitchellh/mapstructure"
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
	Command  *Command    `json:"command,omitempty" yaml:"command,omitempty"`
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
	CopyAction   = "copy"
	OpenAction   = "open"
	PushAction   = "push"
	RunAction    = "run"
	ExitAction   = "exit"
	ReloadAction = "reload"
)

type OnSuccessType string

var (
	OpenOnSuccess   OnSuccessType = "open"
	PushOnSuccess   OnSuccessType = "push"
	ExitOnSuccess   OnSuccessType = "exit"
	ReloadOnSuccess OnSuccessType = "reload"
	CopyOnSuccess   OnSuccessType = "copy"
)

type Action struct {
	Title  string     `json:"title,omitempty" yaml:"title,omitempty"`
	Type   ActionType `json:"type" yaml:"type"`
	Key    string     `json:"key,omitempty" yaml:"key,omitempty"`
	Inputs []Input    `json:"inputs,omitempty" yaml:"inputs,omitempty"`

	// copy
	Text string `json:"text,omitempty" yaml:"text,omitempty"`

	// open
	Target string `json:"target,omitempty" yaml:"target,omitempty"`

	// push
	Page string `json:"page,omitempty" yaml:"page,omitempty"`

	// run
	Command   *Command      `json:"command,omitempty" yaml:"command,omitempty"`
	OnSuccess OnSuccessType `json:"onSuccess,omitempty" yaml:"onSuccess,omitempty"`
}

type Command struct {
	Args  []string `json:"args" yaml:"args"`
	Input string   `json:"input,omitempty" yaml:"input,omitempty"`
	Dir   string   `json:"dir,omitempty" yaml:"dir,omitempty"`
}

func (c Command) Cmd() *exec.Cmd {
	cmd := exec.Command(c.Args[0], c.Args[1:]...)
	cmd.Dir = c.Dir
	cmd.Stdin = strings.NewReader(c.Input)

	return cmd

}

func (c Command) Run() error {
	err := c.Cmd().Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("command exited with %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
	} else if err != nil {
		return err
	}

	return nil

}

func (c Command) Output() ([]byte, error) {
	output, err := c.Cmd().Output()

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return nil, fmt.Errorf("command exited with %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
	} else if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *Command) UnmarshalJSON(data []byte) error {
	var sa []string
	if err := json.Unmarshal(data, &sa); err == nil {
		if len(sa) == 0 {
			return fmt.Errorf("empty command")
		}
		c.Args = sa
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		args, err := shlex.Split(s)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return fmt.Errorf("empty command")
		}

		c.Args = args
		return nil
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err == nil {
		if err := mapstructure.Decode(m, c); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("invalid command")
}

func (c *Command) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sa []string
	if err := unmarshal(&sa); err == nil {
		if len(sa) == 0 {
			return fmt.Errorf("empty command")
		}
		c.Args = sa
		return nil
	}

	var s string
	if err := unmarshal(&s); err == nil {

		args, err := shlex.Split(s)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return fmt.Errorf("empty command")
		}

		c.Args = args
		return nil
	}

	var m map[string]interface{}
	if err := unmarshal(&m); err == nil {
		if err := mapstructure.Decode(m, c); err != nil {
			return err
		}
	}

	return fmt.Errorf("invalid command")
}
