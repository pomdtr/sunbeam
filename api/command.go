package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
)

type SunbeamCommand struct {
	Id       string         `json:"id"`
	Mode     string         `json:"mode"`
	Title    string         `json:"title"`
	Subtitle string         `json:"subtitle"`
	Hidden   bool           `json:"hidden"`
	Params   []SunbeamParam `json:"params"`

	Detail DetailData   `json:"detail"`
	List   ListData     `json:"list"`
	Action ScriptAction `json:"action"`

	ExtensionId string
	Url         url.URL
	Root        url.URL
}

type ScriptAction struct {
	Type    string            `json:"type"`
	Title   string            `json:"title"`
	Path    string            `json:"path"`
	Keybind string            `json:"keybind"`
	Params  map[string]string `json:"params"`
	Target  string            `json:"target,omitempty"`
	CommandData
	Application string `json:"application,omitempty"`
	Url         string `json:"url,omitempty"`
	Content     string `json:"content,omitempty"`
}

type CommandData struct {
	Workdir string `json:"workdir"`
	Command string `json:"command"`
}

type ListData struct {
	CommandData
	ShowDetail bool `json:"showDetail"`
	Dynamic    bool `json:"dynamic"`
}

type DetailData struct {
	CommandData
	Format  string         `json:"format"`
	Text    string         `json:"text"`
	Actions []ScriptAction `json:"actions"`
}

type SunbeamParam struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	Label       string `json:"label"`
	Title       string `json:"title"`
	Optional    bool   `json:"optional"`
	Placeholder string `json:"placeholder"`
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

func (c SunbeamCommand) CheckMissingParams(inputParams map[string]string) []SunbeamParam {
	missing := make([]SunbeamParam, 0)
	for _, param := range c.Params {
		if param.Optional {
			continue
		}
		if _, ok := inputParams[param.Id]; !ok {
			missing = append(missing, param)
		}
	}
	return missing
}

func (l ListData) Run(input CommandInput) ([]ScriptItem, error) {
	output, err := l.CommandData.Run(input)
	if err != nil {
		return nil, err
	}
	rows := strings.Split(output, "\n")
	items := make([]ScriptItem, 0)
	for _, row := range rows {
		if row == "" {
			continue
		}
		var item ScriptItem
		err := json.Unmarshal([]byte(row), &item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (c SunbeamCommand) Target() string {
	return fmt.Sprintf("%s/%s", c.ExtensionId, c.Id)
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

func (c CommandData) Run(input CommandInput) (string, error) {
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

	log.Printf("Executing command: %s", rendered)
	cmd := exec.Command("sh", "-c", rendered)
	cmd.Dir = c.Workdir

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
	Fill     string         `json:"fill"`
	Actions  []ScriptAction `json:"actions"`
}
