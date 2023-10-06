package tui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

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

func (e Extension) Run(input types.CommandInput) ([]byte, error) {
	cmd, err := e.Cmd(input)
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

func (e Extension) Cmd(input types.CommandInput) (*exec.Cmd, error) {
	if input.Params == nil {
		input.Params = make(map[string]any)
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	command := exec.Command(e.Entrypoint, string(inputBytes))
	command.Env = os.Environ()
	command.Env = append(command.Env, "SUNBEAM=0")
	command.Env = append(command.Env, "NO_COLOR=1")

	return command, nil
}

type Extension struct {
	types.Manifest
	Entrypoint string
}

func LoadExtension(entrypoint string) (Extension, error) {
	command := exec.Command(entrypoint)
	b, err := command.Output()
	if err != nil {
		return Extension{}, err
	}

	if err := schemas.ValidateManifest(b); err != nil {
		return Extension{}, err
	}

	var manifest types.Manifest
	if err := json.Unmarshal(b, &manifest); err != nil {
		return Extension{}, err
	}

	return Extension{
		Manifest:   manifest,
		Entrypoint: entrypoint,
	}, nil
}
