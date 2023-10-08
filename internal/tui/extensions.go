package tui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	command := exec.Command(e.Entrypoint, commandName)
	command.Stdin = strings.NewReader(string(inputBytes))
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
	b, err := exec.Command(entrypoint).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return Extension{}, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
		}

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
