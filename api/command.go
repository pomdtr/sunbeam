package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
)

type SunbeamCommand struct {
	Type    string     `json:"type"`
	Params  []FormItem `json:"params"`
	Command string     `json:"command"`
	DetailParam
	ListParam

	Extension string
	Url       url.URL
	Root      url.URL
}

type ListParam struct {
	ShowDetail bool `json:"showDetail"`
	Dynamic    bool `json:"dynamic"`
}

type DetailParam struct {
	Format string `json:"format"`
}

type DetailData struct {
	DetailParam
	Command string `json:"command"`
}

type FormItem struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Placeholder string `json:"placeholder"`
	Label       string `json:"label"`
}

type CommandInput struct {
	Params map[string]string
	Query  string
}

func NewCommandInput(params map[string]string) CommandInput {
	if params == nil {
		params = make(map[string]string)
	}
	return CommandInput{Params: params}
}

func (c SunbeamCommand) CheckMissingParams(inputParams map[string]string) []FormItem {
	missing := make([]FormItem, 0)
	for _, param := range c.Params {
		if _, ok := inputParams[param.Name]; !ok {
			missing = append(missing, param)
		}
	}
	return missing
}

func renderCommand(command string, data map[string]any) (string, error) {
	t, err := template.New("").Parse(command)
	if err != nil {
		return "", err
	}

	out := bytes.Buffer{}
	if err = t.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (c SunbeamCommand) Run(input CommandInput) (string, error) {
	var err error
	params := make(map[string]any)
	for key, value := range input.Params {
		params[key] = shellescape.Quote(value)
	}

	rendered, err := renderCommand(c.Command, map[string]any{
		"params": params,
		"query":  shellescape.Quote(input.Query),
	})
	if err != nil {
		return "", err
	}

	cmd := exec.Command("sh", "-c", rendered)
	cmd.Dir = c.Root.Path

	var outbuf, errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error while running command: %s", errbuf.String())
	}
	return outbuf.String(), nil
}

func (c SunbeamCommand) RemoteRun(input CommandInput) (string, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	res, err := http.Post(http.MethodPost, c.Url.String(), bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	bytes, _ := io.ReadAll(res.Body)

	return string(bytes), nil
}

type ScriptItem struct {
	Icon     string         `json:"icon"`
	Title    string         `json:"title"`
	Subtitle string         `json:"subtitle"`
	Detail   DetailData     `json:"detail"`
	Actions  []ScriptAction `json:"actions"`
}

type ScriptAction struct {
	Type        string            `json:"type"`
	Shortcut    string            `json:"shortcut"`
	Title       string            `json:"title"`
	Path        string            `json:"path"`
	Params      map[string]string `json:"params"`
	Extension   string            `json:"extension"`
	Target      string            `json:"target,omitempty"`
	Command     string            `json:"command,omitempty"`
	Application string            `json:"application,omitempty"`
	Url         string            `json:"url,omitempty"`
	Content     string            `json:"content,omitempty"`
}

func ParseScriptItems(output string) (items []ScriptItem, err error) {
	rows := strings.Split(output, "\n")
	for _, row := range rows {
		if row == "" {
			continue
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
