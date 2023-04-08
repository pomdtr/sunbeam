package utils

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/google/shlex"
)

func RunCommand(command string, dir string) ([]byte, error) {
	args, err := shlex.Split(command)
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	var cmd *exec.Cmd
	if len(args) == 1 {
		cmd = exec.Command(args[0])
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return nil, fmt.Errorf("command exit with code %d: %s", exitError.ExitCode(), exitError.Stderr)
		}

		return nil, err
	}

	return output, nil

}
