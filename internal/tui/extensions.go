package tui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
)

func (e Extension) Command(name string) (types.CommandSpec, bool) {
	for _, command := range e.Commands {
		if command.Name == name {
			return command, true
		}
	}
	return types.CommandSpec{}, false
}

func (e Extension) Run(command string, input types.CommandInput) ([]byte, error) {
	cmd, err := e.Cmd(command, input)
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

	workdir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	input.WorkDir = workdir

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

type Extension struct {
	types.Manifest
	Entrypoint string
}

func LoadExtension(entrypoint string) (Extension, error) {
	cmd := exec.Command(entrypoint)
	cmd.Dir = filepath.Dir(entrypoint)

	manifestBytes, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return Extension{}, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
		}

		return Extension{}, err
	}

	if err := schemas.ValidateManifest(manifestBytes); err != nil {
		return Extension{}, err
	}

	var manifest types.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return Extension{}, err
	}

	return Extension{
		Manifest:   manifest,
		Entrypoint: entrypoint,
	}, nil
}
