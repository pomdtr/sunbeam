package extensions

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type ExtensionMap map[string]Extension

type Extension struct {
	types.Manifest
	Entrypoint string
}

type Metadata struct {
	Type       ExtensionType `json:"type"`
	Origin     string        `json:"origin"`
	Entrypoint string        `json:"entrypoint"`
}

type ExtensionType string

const (
	ExtensionTypeLocal ExtensionType = "local"
	ExtensionTypeGit   ExtensionType = "git"
	ExtensionTypeHttp  ExtensionType = "http"
)

func IsRootCommand(command types.CommandSpec) bool {
	if command.Hidden {
		return false
	}

	for _, param := range command.Params {
		if param.Required {
			return false
		}
	}

	return true
}

func (e Extension) Command(name string) (types.CommandSpec, bool) {
	for _, command := range e.Commands {
		if command.Name == name {
			return command, true
		}
	}
	return types.CommandSpec{}, false
}

func (e Extension) RootCommands() []types.CommandSpec {
	rootCommands := make([]types.CommandSpec, 0)
	for _, command := range e.Commands {
		if !IsRootCommand(command) {
			continue
		}

		rootCommands = append(rootCommands, command)
	}

	return rootCommands
}

func (ext Extension) Run(command string, input types.CommandInput) ([]byte, error) {
	cmd, err := ext.Cmd(command, input)
	if err != nil {
		return nil, err
	}

	var exitErr *exec.ExitError
	if output, err := cmd.Output(); err == nil {
		return output, nil
	} else if errors.As(err, &exitErr) {
		return nil, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
	} else {
		return nil, err
	}
}

func (e Extension) Cmd(commandName string, input types.CommandInput) (*exec.Cmd, error) {
	if input.Params == nil {
		input.Params = make(map[string]any)
	}

	command, ok := e.Command(commandName)
	if !ok {
		return nil, fmt.Errorf("command %s not found", commandName)
	}

	for _, spec := range command.Params {
		_, ok := input.Params[spec.Name]
		if !ok && spec.Required {
			return nil, fmt.Errorf("missing required parameter %s", spec.Name)
		}

		if spec.Default != nil {
			input.Params[spec.Name] = spec.Default
		}
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(e.Entrypoint, commandName)
	cmd.Stdin = strings.NewReader(string(inputBytes))
	cmd.Dir = filepath.Dir(e.Entrypoint)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SUNBEAM=0")
	cmd.Env = append(cmd.Env, "NO_COLOR=1")

	return cmd, nil
}

func HasMissingParams(command types.CommandSpec, params map[string]any) bool {
	return len(FindMissingParams(command, params)) > 0
}

func FindMissingParams(command types.CommandSpec, params map[string]any) []types.Param {
	missing := make([]types.Param, 0)
	for _, spec := range command.Params {
		if !spec.Required {
			continue
		}

		_, ok := params[spec.Name]
		if ok {
			continue
		}

		missing = append(missing, spec)
	}

	return missing
}
