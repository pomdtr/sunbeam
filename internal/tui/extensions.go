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

type CommandInput struct {
	Command string         `json:"command"`
	Params  map[string]any `json:"params"`
	Inputs  map[string]any `json:"inputs,omitempty"`
	Query   string         `json:"query,omitempty"`
}

func (e Extension) Command(name string) (types.CommandSpec, bool) {
	for _, command := range e.Commands {
		if command.Name == name {
			return command, true
		}
	}
	return types.CommandSpec{}, false
}

func ShellCommand(ref CommandRef) string {
	args := []string{"sunbeam", "run", ref.Script, ref.Command}
	for name, value := range ref.Params {
		switch value := value.(type) {
		case string:
			args = append(args, fmt.Sprintf("--%s=%s", name, value))
		case bool:
			if value {
				args = append(args, fmt.Sprintf("--%s", name))
			}
		}
	}

	return strings.Join(args, " ")
}

func (e Extension) Run(input CommandInput) ([]byte, error) {
	if input.Params == nil {
		input.Params = make(map[string]any)
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	command := exec.Command(e.Path, string(inputBytes))
	command.Env = os.Environ()
	command.Env = append(command.Env, "NO_COLOR=1")

	var exitErr *exec.ExitError
	if output, err := command.Output(); err == nil {
		return output, nil
	} else if errors.As(err, &exitErr) {
		return nil, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
	} else {
		return nil, err
	}
}

type Extension struct {
	types.Manifest
	Path string
}

type Extensions map[string]Extension

func (extensions Extensions) Get(extensionpath string) (Extension, error) {
	extensionpath, err := filepath.Abs(extensionpath)
	if err != nil {
		return Extension{}, err
	}

	if extension, ok := extensions[extensionpath]; ok {
		return extension, nil
	}

	extension, err := LoadExtension(extensionpath)
	if err != nil {
		return extension, err
	}

	if extensions == nil {
		extensions = make(map[string]Extension)
	}
	extensions[extensionpath] = extension

	return extension, nil
}

func LoadExtension(extensionpath string) (Extension, error) {
	command := exec.Command(extensionpath)
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
		Manifest: manifest,
		Path:     extensionpath,
	}, nil
}
