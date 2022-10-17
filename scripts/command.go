package scripts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/adrg/xdg"
)

var CommandDir string

func init() {
	if commandDir := os.Getenv("SUNBEAM_SCRIPT_DIR"); commandDir != "" {
		CommandDir = commandDir
	} else {
		CommandDir = xdg.DataHome
	}
}

type Command struct {
	Script
	CommandInput
}

type CommandInput struct {
	Environment map[string]string `json:"environment"`
	Arguments   []string          `json:"arguments"`
	Query       string            `json:"query"`
}

func (c Command) Run() (res ScriptResponse, err error) {
	log.Printf("Running command %s with args %s", c.Script.Url.Path, c.Arguments)
	// Check if the number of arguments is correct
	if c.Url.Scheme != "file" {
		payload, err := json.Marshal(c.CommandInput)
		if err != nil {
			return res, err
		}

		httpRes, err := http.Post(http.MethodPost, c.Url.String(), bytes.NewBuffer(payload))
		if err != nil {
			return ScriptResponse{
				Type: "detail",
				Detail: &DetailResponse{
					Format: "text",
					Text:   fmt.Errorf("error while running command: %s", err).Error(),
				},
			}, nil
		}
		var res ScriptResponse
		err = json.NewDecoder(httpRes.Body).Decode(&res)
		if err != nil {
			log.Fatalf("Error while decoding response: %s", err)
		}

		return res, nil
	}
	if len(c.Arguments) < len(c.RequiredArguments()) {
		formItems := make([]FormItem, 0)
		for i := len(c.Arguments); i < len(c.Metadatas.Arguments); i++ {
			formItems = append(formItems, FormItem{
				Type: "text",
				Id:   c.Metadatas.Arguments[i].Placeholder,
				Name: c.Metadatas.Arguments[i].Placeholder,
			})
		}
		return ScriptResponse{
			Type: "form",
			Form: &FormResponse{
				Method: "args",
				Items:  formItems,
			},
		}, nil
	}

	cmd := exec.Command(c.Script.Url.Path, c.Arguments...)
	cmd.Dir = path.Dir(cmd.Path)

	// Copy process environment
	cmd.Env = make([]string, len(os.Environ()))
	copy(cmd.Env, os.Environ())

	// Add custom environment variables to the process
	for key, value := range c.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	var inbuf, outbuf, errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf
	cmd.Stdin = &inbuf

	if c.Metadatas.Mode == "interactive" {
		// Add support dir to environment
		supportDir := path.Join(xdg.DataHome, "sunbeam", c.Script.Metadatas.PackageName, "support")
		cmd.Env = append(cmd.Env, fmt.Sprintf("SUNBEAM_SUPPORT_DIR=%s", supportDir))

		inbuf.Write([]byte(c.Query))
	}

	err = cmd.Run()

	if err != nil {
		return ScriptResponse{
			Type: "detail",
			Detail: &DetailResponse{
				Format: "text",
				Text:   errbuf.String(),
			},
		}, nil
	}

	if c.Metadatas.Mode != "interactive" {
		return ScriptResponse{
			Type: "detail",
			Detail: &DetailResponse{
				Format: "text",
				Text:   outbuf.String(),
			},
		}, nil
	}

	err = json.Unmarshal(outbuf.Bytes(), &res)
	if err != nil {
		return
	}
	err = Validator.Struct(res)
	if err != nil {
		return ScriptResponse{
			Type: "detail",
			Detail: &DetailResponse{
				Format: "text",
				Actions: []ScriptAction{
					{Type: "copy", Content: outbuf.String()},
				},
				Text: err.Error(),
			},
		}, nil
	}

	return
}
