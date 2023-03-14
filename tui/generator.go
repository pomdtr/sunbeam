package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pomdtr/sunbeam/types"

	"gopkg.in/yaml.v3"
)

type PageGenerator func(input string) ([]byte, error)

type CmdGenerator struct {
	Command string
	Args    []string
	Dir     string
}

func NewCommandGenerator(command string, args []string, dir string) PageGenerator {
	return func(input string) ([]byte, error) {
		command := exec.Command(command, args...)
		command.Stdin = strings.NewReader(input)
		command.Dir = dir
		output, err := command.Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				return nil, fmt.Errorf("script exited with code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
			}

			return nil, err
		}

		return output, nil
	}
}

func NewFileGenerator(name string) PageGenerator {
	return func(input string) ([]byte, error) {
		if path.Ext(name) == ".json" {
			return os.ReadFile(name)
		}

		if path.Ext(name) == ".yaml" || path.Ext(name) == ".yml" {
			bytes, err := os.ReadFile(name)
			if err != nil {
				return nil, err
			}

			var page types.Page
			if err := yaml.Unmarshal(bytes, &page); err != nil {
				return nil, err
			}

			return json.Marshal(page)
		}

		return nil, fmt.Errorf("unsupported file type")
	}
}
