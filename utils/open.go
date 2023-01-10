package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

func Open(target string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		command = exec.Command("xdg-open", target)
	case "darwin":
		command = exec.Command("open", target)
	}

	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to open %s: %s", target, output)
	}
	return nil
}
