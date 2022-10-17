package scripts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

type Command interface {
	Run(CommandInput) (ScriptResponse, error)
	Title() string
}

type CommandInput struct {
	Environment map[string]string `json:"environment"`
	Arguments   []string          `json:"arguments"`
	Query       string            `json:"query"`
}

type RemoteCommand struct {
	Url   url.URL
	Title string
}

type RootCommand struct {
	Root string
}

func (c RootCommand) Title() string {
	return "Commands"
}

func (c RootCommand) Run(input CommandInput) (res ScriptResponse, err error) {
	dirCommands, err := ScanDir(c.Root)
	if err != nil {
		return
	}

	items := make([]ScriptItem, len(dirCommands))
	for i, command := range dirCommands {
		items[i] = ScriptItem{
			RawTitle: command.Title(),
			Actions: []ScriptAction{
				{
					Type: "push",
					Path: command.Path,
				},
			},
		}
	}

	res.Type = "list"
	res.List = &ListResponse{
		Items: items,
	}

	return
}

func (c RemoteCommand) Run(input CommandInput) (res ScriptResponse, err error) {
	payload, err := json.Marshal(input)
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

	err = json.NewDecoder(httpRes.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("error while decoding response: %s", err)
	}

	return res, nil
}

func (c LocalCommand) Run(input CommandInput) (res ScriptResponse, err error) {
	log.Printf("Running command %s with args %s", c.Path, input.Arguments)
	// Check if the number of arguments is correct
	if len(input.Arguments) < len(c.RequiredArguments()) {
		formItems := make([]FormItem, 0)
		for i := len(input.Arguments); i < len(c.ScriptMetadatas.Arguments); i++ {
			formItems = append(formItems, FormItem{
				Type: "text",
				Id:   c.ScriptMetadatas.Arguments[i].Placeholder,
				Name: c.ScriptMetadatas.Arguments[i].Placeholder,
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

	cmd := exec.Command(c.Path, input.Arguments...)
	cmd.Dir = path.Dir(cmd.Path)

	// Copy process environment
	cmd.Env = make([]string, len(os.Environ()))
	copy(cmd.Env, os.Environ())

	// Add custom environment variables to the process
	for key, value := range input.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	var inbuf, outbuf, errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf
	cmd.Stdin = &inbuf

	if c.ScriptMetadatas.Mode == "interactive" {
		// Add support dir to environment
		supportDir := path.Join(xdg.DataHome, "sunbeam", c.ScriptMetadatas.PackageName, "support")
		cmd.Env = append(cmd.Env, fmt.Sprintf("SUNBEAM_SUPPORT_DIR=%s", supportDir))

		inbuf.Write([]byte(input.Query))
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

	if c.ScriptMetadatas.Mode != "interactive" {
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
