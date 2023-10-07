package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func EditCmd(fp string) *exec.Cmd {
	return exec.Command("sh", "-c", fmt.Sprintf("%s %s", editor(), fp))
}

func editor() string {
	if editor, ok := os.LookupEnv("VISUAL"); ok {
		return editor
	}

	if editor, ok := os.LookupEnv("EDITOR"); ok {
		return editor
	}

	return "vim"
}
