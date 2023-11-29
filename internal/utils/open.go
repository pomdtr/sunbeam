package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

func Open(target string) error {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open", target}
	case "linux":
		args = []string{"xdg-open", target}
	default:
		return fmt.Errorf("unsupportedPlatform")
	}

	return runDetached(args...)
}

func runDetached(args ...string) error {
	cmd := exec.Command("nohup", args...)

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Process.Release()
}
