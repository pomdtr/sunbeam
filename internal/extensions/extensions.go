package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/acarl005/stripansi"
	"github.com/cli/cli/pkg/findsh"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type ExtensionMap map[string]Extension

func (e ExtensionMap) List() []Extension {
	extensions := make([]Extension, 0)
	for _, extension := range e {
		extensions = append(extensions, extension)
	}
	return extensions
}

type Extension struct {
	types.Manifest `json:"manifest"`
	Metadata
}

type Preferences map[string]any

type Metadata struct {
	Type       ExtensionType `json:"type"`
	Origin     string        `json:"origin"`
	Entrypoint string        `json:"entrypoint"`
}

type ExtensionType string

const (
	ExtensionTypeLocal ExtensionType = "local"
	ExtensionTypeHttp  ExtensionType = "http"
)

func (e Extension) Command(name string) (types.CommandSpec, bool) {
	for _, command := range e.Commands {
		if command.Name == name {
			return command, true
		}
	}
	return types.CommandSpec{}, false
}

func (e Extension) RootItems() []types.RootItem {
	rootItems := make([]types.RootItem, 0)
	if e.Root != nil {
		for _, rootItem := range e.Root {
			command, ok := e.Command(rootItem.Command)
			if !ok {
				continue
			}

			if rootItem.Title == "" {
				rootItem.Title = command.Title
			}

			rootItems = append(rootItems, rootItem)
		}
	} else {
		for _, command := range e.Commands {
			if command.Hidden {
				continue
			}

			rootItems = append(rootItems, types.RootItem{
				Title:   command.Title,
				Command: command.Name,
			})
		}
	}

	return rootItems
}

func (e Extension) Run(input types.CommandInput) error {
	_, err := e.Output(input)
	return err
}

func (ext Extension) CheckRequirements() error {
	for _, requirement := range ext.Require {
		if _, err := exec.LookPath(requirement.Name); err != nil {
			return fmt.Errorf("missing requirement %s", requirement.Name)
		}
	}

	return nil
}

func (ext Extension) Output(input types.CommandInput) ([]byte, error) {
	cmd, err := ext.Cmd(input)
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
	return e.CmdContext(context.Background(), input)
}

func (e Extension) CmdContext(ctx context.Context, input types.CommandInput) (*exec.Cmd, error) {
	if input.Params == nil {
		input.Params = make(map[string]any)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	input.Cwd = cwd

	command, ok := e.Command(input.Command)
	if !ok {
		return nil, fmt.Errorf("command %s not found", input.Command)
	}

	for _, spec := range command.Inputs {
		if !spec.Required {
			if spec.Default != nil {
				input.Params[spec.Name] = spec.Default
			}

			continue
		}
		_, ok := input.Params[spec.Name]
		if !ok {
			return nil, fmt.Errorf("missing required parameter %s", spec.Name)
		}
	}

	for _, spec := range e.Preferences {
		if !spec.Required {
			if spec.Default != nil {
				input.Params[spec.Name] = spec.Default
			}

			continue
		}
		_, ok := input.Preferences[spec.Name]
		if !ok {
			return nil, fmt.Errorf("missing required preference %s", spec.Name)
		}
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var args []string
	if runtime.GOOS == "windows" {
		sh, err := findsh.Find()
		if err != nil {
			return nil, err
		}
		args = []string{sh, "-s", "-c", `command "$@"`, "--", e.Entrypoint, string(inputBytes)}
	} else {
		args = []string{e.Entrypoint, string(inputBytes)}
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = filepath.Dir(e.Entrypoint)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SUNBEAM=1")
	return cmd, nil
}
