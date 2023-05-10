package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/shlex"
	"github.com/mitchellh/mapstructure"
)

type PageType string

const (
	DetailPage PageType = "detail"
	ListPage   PageType = "list"
)

type Page struct {
	Type    PageType `json:"type"`
	Title   string   `json:"title,omitempty"`
	Actions []Action `json:"actions,omitempty"`

	// Detail page
	Preview *TextOrCommand `json:"preview,omitempty"`

	// List page
	ShowPreview   bool       `json:"showPreview,omitempty"`
	OnQueryChange *Command   `json:"onQueryChange,omitempty"`
	EmptyView     *EmptyView `json:"emptyView,omitempty"`
	Items         []ListItem `json:"items,omitempty"`
}

type EmptyView struct {
	Text    string   `json:"text,omitempty"`
	Actions []Action `json:"actions,omitempty"`
}

type ListItem struct {
	Id          string         `json:"id,omitempty"`
	Title       string         `json:"title"`
	Subtitle    string         `json:"subtitle,omitempty"`
	Preview     *TextOrCommand `json:"preview,omitempty"`
	Accessories []string       `json:"accessories,omitempty"`
	Actions     []Action       `json:"actions,omitempty"`
}

type FormInputType string

const (
	TextFieldInput FormInputType = "textfield"
	TextAreaInput  FormInputType = "textarea"
	DropDownInput  FormInputType = "dropdown"
	CheckboxInput  FormInputType = "checkbox"
)

type DropDownItem struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type Input struct {
	Name        string        `json:"name"`
	Type        FormInputType `json:"type"`
	Title       string        `json:"title"`
	Placeholder string        `json:"placeholder,omitempty"`
	Default     any           `json:"default,omitempty"`
	Optional    bool          `json:"optional,omitempty"`

	// Only for dropdown
	Items []DropDownItem `json:"items,omitempty"`

	// Only for checkbox
	Label             string `json:"label,omitempty"`
	TrueSubstitution  string `json:"trueSubstitution,omitempty"`
	FalseSubstitution string `json:"falseSubstitution,omitempty"`
}

func NewTextInput(name string, title string, placeholder string) Input {
	return Input{
		Name:        name,
		Type:        TextFieldInput,
		Title:       title,
		Placeholder: placeholder,
	}
}

func NewTextAreaInput(name string, title string, placeholder string) Input {
	return Input{
		Name:        name,
		Type:        TextAreaInput,
		Title:       title,
		Placeholder: placeholder,
	}
}

func NewCheckbox(name string, title string, label string) Input {
	return Input{
		Name:  name,
		Type:  CheckboxInput,
		Title: title,
		Label: label,
	}
}

func NewDropDown(name string, title string, items ...DropDownItem) Input {
	return Input{
		Name:  name,
		Type:  DropDownInput,
		Title: title,
		Items: items,
	}
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

type Action struct {
	Title  string     `json:"title,omitempty"`
	Type   ActionType `json:"type"`
	Key    string     `json:"key,omitempty"`
	Inputs []Input    `json:"inputs,omitempty"`

	// copy
	Text string `json:"text,omitempty"`

	// open
	Target string `json:"target,omitempty"`

	// push
	Page *TextOrCommand `json:"page,omitempty"`

	// run
	Command         *Command `json:"command,omitempty"`
	ReloadOnSuccess bool     `json:"reloadOnSuccess,omitempty"`
}

type TextOrCommand struct {
	Text    string   `json:"text,omitempty"`
	Command *Command `json:"command,omitempty"`
}

func (p *TextOrCommand) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		p.Text = text
		return nil
	}

	var v struct {
		Text    string   `json:"text"`
		Command *Command `json:"command"`
	}

	if err := json.Unmarshal(data, &v); err == nil {
		p.Text = v.Text
		p.Command = v.Command
		return nil
	}

	return errors.New("page must be a string or a command")
}

func NewReloadAction() Action {
	return Action{
		Type: ReloadAction,
	}
}

func NewCopyAction(title string, text string) Action {
	return Action{
		Title: title,
		Type:  CopyAction,
		Text:  text,
	}
}

func NewOpenAction(title string, target string) Action {
	return Action{
		Title:  title,
		Type:   OpenAction,
		Target: target,
	}
}

func NewPushAction(title string, page string) Action {
	return Action{
		Title: title,
		Type:  PushAction,
		Page: &TextOrCommand{
			Text: page,
		},
	}
}

func NewRunAction(title string, name string, args ...string) Action {
	return Action{
		Title: title,
		Type:  RunAction,
		Command: &Command{
			Args: args,
		},
	}
}

type Command struct {
	Name  string   `json:"name"`
	Args  []string `json:"args,omitempty"`
	Input string   `json:"input,omitempty"`
	Dir   string   `json:"dir,omitempty"`
}

func (c Command) Cmd() *exec.Cmd {
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Dir = c.Dir
	if c.Input != "" {
		cmd.Stdin = strings.NewReader(c.Input)
	}

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
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		args, err := shlex.Split(s)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return fmt.Errorf("empty command")
		}

		c.Name = args[0]
		c.Args = args[1:]
		return nil
	}

	var args []string
	if err := json.Unmarshal(data, &args); err == nil {
		if len(args) == 0 {
			return fmt.Errorf("empty command")
		}
		c.Name = args[0]
		c.Args = args[1:]
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
