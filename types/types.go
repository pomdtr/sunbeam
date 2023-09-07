package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cli/cli/v2/pkg/findsh"
	"github.com/mitchellh/mapstructure"
)

type List struct {
	Title     string     `json:"title,omitempty"`
	EmptyView *EmptyView `json:"emptyView,omitempty"`
	Items     []ListItem `json:"items"`
}

type Detail struct {
	Title     string   `json:"title,omitempty"`
	Actions   []Action `json:"actions,omitempty"`
	Highlight string   `json:"highlight,omitempty"`
	Text      string   `json:"text"`
}

type Form struct {
	Title        string  `json:"title,omitempty"`
	SubmitAction *Action `json:"submitAction,omitempty"`
}

type EmptyView struct {
	Text    string   `json:"text,omitempty"`
	Actions []Action `json:"actions,omitempty"`
}

type ListItem struct {
	Id          string   `json:"id,omitempty"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle,omitempty"`
	Accessories []string `json:"accessories,omitempty"`
	Actions     []Action `json:"actions,omitempty"`
}

type FormInputType string

const (
	TextInput     FormInputType = "text"
	TextAreaInput FormInputType = "textarea"
	SelectInput   FormInputType = "select"
	CheckboxInput FormInputType = "checkbox"
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
		Type:        TextInput,
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
		Type:  SelectInput,
		Title: title,
		Items: items,
	}
}

type ActionType string

const (
	CopyAction   = "copy"
	OpenAction   = "open"
	RunAction    = "run"
	ShareAction  = "share"
	PasteAction  = "paste"
	ReloadAction = "reload"
)

type Action struct {
	Title string     `json:"title,omitempty"`
	Type  ActionType `json:"type"`
	Key   string     `json:"key,omitempty"`

	// copy
	Text string `json:"text,omitempty"`

	// open
	Target string `json:"target,omitempty"`

	// push
	Page string `json:"page,omitempty"`

	// run
	Command *Command `json:"command,omitempty"`

	// share
	Params Params `json:"params,omitempty"`
}

type Params struct {
	Text  string `json:"text,omitempty"`
	Title string `json:"title,omitempty"`
	Url   string `json:"url,omitempty"`
}

func (a Action) Output(ctx context.Context) ([]byte, error) {
	if a.Command != nil {
		return a.Command.Output(ctx)
	} else {
		return nil, errors.New("invalid action")
	}
}

type Request struct {
	Url     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

func (r *Request) UnmarshalJSON(data []byte) error {
	var url string
	if err := json.Unmarshal(data, &url); err == nil {
		r.Url = url
		return nil
	}

	var request map[string]any
	if err := json.Unmarshal(data, &request); err == nil {
		if err := mapstructure.Decode(request, r); err != nil {
			return err
		}

		return nil
	}

	return errors.New("invalid request")
}

func (r Request) Do(ctx context.Context) ([]byte, error) {
	if r.Method == "" {
		r.Method = http.MethodGet
	}

	req, err := http.NewRequest(r.Method, r.Url, strings.NewReader(r.Body))
	if err != nil {
		return nil, err
	}

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New(resp.Status)
	}

	return io.ReadAll(resp.Body)
}

type Command struct {
	Name  string   `json:"name"`
	Args  []string `json:"args,omitempty"`
	Input string   `json:"input,omitempty"`
}

func (c Command) Cmd(ctx context.Context) (*exec.Cmd, error) {
	var name string
	var args []string

	if runtime.GOOS != "windows" {
		shExe, err := findsh.Find()
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				return nil, errors.New("the `sh.exe` interpreter is required. Please install Git for Windows and try again")
			}

			return nil, err
		}
		name = shExe
		args = append(args, "-c", `command "$@"`, "--", c.Name)
		args = append(args, c.Args...)

	} else {
		name = c.Name
		args = append(args, c.Args...)
	}

	cmd := exec.CommandContext(ctx, name, args...)
	if c.Input != "" {
		cmd.Stdin = strings.NewReader(c.Input)
	}

	return cmd, nil

}

func (c Command) Run(ctx context.Context) error {
	cmd, err := c.Cmd(ctx)
	if err != nil {
		return err
	}

	var exitErr *exec.ExitError
	if err := cmd.Run(); errors.As(err, &exitErr) {
		return fmt.Errorf("command exited with %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
	} else if err != nil {
		return err
	}

	return nil
}

func (c Command) Output(ctx context.Context) ([]byte, error) {
	cmd, err := c.Cmd(ctx)
	if err != nil {
		return nil, err
	}

	output, err := cmd.Output()

	var exitErr *exec.ExitError
	var pathErr *fs.PathError
	if errors.As(err, &exitErr) {
		return nil, fmt.Errorf("command exited with %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))

	} else if errors.As(err, &pathErr) {
		if strings.Contains(err.Error(), "permission denied") && runtime.GOOS != "windows" {
			return nil, fmt.Errorf("permission denied, try running `chmod +x %s`", c.Name)
		}
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("command failed (%s): %s", cmd.String(), err)
	}

	return output, nil
}

func (c *Command) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		c.Name = "bash"
		c.Args = []string{"-c", s}
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
