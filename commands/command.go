package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"text/template"
)

type CommandData struct {
	Id       string         `json:"id"`
	Title    string         `json:"title"`
	Command  string         `json:"command"`
	Subtitle string         `json:"subtitle"`
	Mode     string         `json:"mode"`
	Params   []CommandParam `json:"params"`
}

type CommandParam struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	Label       string `json:"label"`
	Title       string `json:"title"`
	Optional    bool   `json:"optional"`
	Placeholder string `json:"placeholder"`
}

type CommandInput struct {
	Environment map[string]string `json:"environment"`
	Params      map[string]any    `json:"params"`
	Query       string            `json:"query"`
}

type Command struct {
	CommandData
	Root url.URL
	Url  url.URL
}

func (c Command) CheckMissingParams(inputParams map[string]any) []CommandParam {
	missing := make([]CommandParam, 0)
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

func (c Command) Run(input CommandInput) (string, error) {
	switch c.Url.Scheme {
	case "file":
		return c.LocalRun(input)
	case "http", "https":
		return c.RemoteRun(input)
	default:
		return "", fmt.Errorf("unknown command scheme: %s", c.Root.Scheme)
	}
}

func renderCommand(command string, data any) (string, error) {
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

func (c Command) LocalRun(input CommandInput) (string, error) {
	var err error
	rendered, err := renderCommand(c.Command, input)
	if err != nil {
		return "", err
	}

	log.Printf("Executing command: %s", rendered)
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

func (c Command) RemoteRun(input CommandInput) (string, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	res, err := http.Post(http.MethodPost, c.Url.String(), bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	bytes, err := io.ReadAll(res.Body)

	return string(bytes), nil
}
