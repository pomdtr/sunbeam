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

type SunbeamScript struct {
	Output  string     `json:"output"`
	Params  []FormItem `json:"params"`
	Command string     `json:"command"`

	// Detail Properties
	Format string `json:"format"`
	// Remote Properties
	Dynamic bool `json:"dynamic"`

	Extension string
	Url       url.URL
	Root      url.URL
}

type FormItem struct {
	Type        string         `json:"type"`
	Id          string         `json:"id"`
	Title       string         `json:"title"`
	Placeholder string         `json:"placeholder"`
	Label       string         `json:"label"`
	Data        []DropDownItem `json:"data"`
}

type DropDownItem struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type ScriptInput struct {
	Params map[string]string
	Query  string
}

func NewScriptInput(params map[string]string) ScriptInput {
	if params == nil {
		params = make(map[string]string)
	}
	return ScriptInput{Params: params}
}

func (s SunbeamScript) CheckMissingParams(inputParams map[string]string) []FormItem {
	missing := make([]FormItem, 0)
	for _, param := range s.Params {
		if _, ok := inputParams[param.Id]; !ok {
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

func (s SunbeamScript) Run(input ScriptInput) (string, error) {
	var err error
	params := make(map[string]any)
	for key, value := range input.Params {
		params[key] = shellescape.Quote(value)
	}

	rendered, err := renderCommand(s.Command, map[string]any{
		"params": params,
		"query":  shellescape.Quote(input.Query),
	})
	if err != nil {
		return "", err
	}

	cmd := exec.Command("sh", "-c", rendered)
	cmd.Dir = s.Root.Path

	var outbuf, errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error while running command: %s", errbuf.String())
	}
	return outbuf.String(), nil
}

func (s SunbeamScript) RemoteRun(input ScriptInput) (string, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	res, err := http.Post(http.MethodPost, s.Url.String(), bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	bytes, _ := io.ReadAll(res.Body)

	return string(bytes), nil
}

type ListItem struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Detail   struct {
		Command string `json:"command"`
		Format  string `json:"format"`
	} `json:"detail"`
	Actions     []ScriptAction `json:"actions"`
	Accessories []string       `json:"accessories"`

	Extension string
}

type ScriptAction struct {
	Title    string `json:"title"`
	Type     string `json:"type"`
	Shortcut string `json:"shortcut"`

	Command string `json:"command,omitempty"`

	Target    string            `json:"target,omitempty"`
	Extension string            `json:"extension,omitempty"`
	Params    map[string]string `json:"params"`

	Url         string `json:"url,omitempty"`
	Path        string `json:"path"`
	Application string `json:"application,omitempty"`

	Content string `json:"content,omitempty"`
}

func ParseListItems(output string) (items []ListItem, err error) {
	rows := strings.Split(output, "\n")
	for _, row := range rows {
		if row == "" {
			continue
		}
		var item ListItem
		err = json.Unmarshal([]byte(row), &item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
