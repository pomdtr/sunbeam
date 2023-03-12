package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
				return nil, fmt.Errorf("Script exited with code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
			}

			return nil, err
		}

		return output, nil
	}
}

func NewFileGenerator(path string) PageGenerator {
	return func(input string) ([]byte, error) {
		return os.ReadFile(path)
	}
}

func NewStaticGenerator(content []byte) PageGenerator {
	return func(input string) ([]byte, error) {
		return content, nil
	}
}
