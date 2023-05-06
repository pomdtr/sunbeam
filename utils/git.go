package utils

import (
	"os"
	"os/exec"
)

func GitClone(repoUrl string, targetDir string) error {
	command := exec.Command("git", "clone", repoUrl, targetDir)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}

func GitPull(localDir string) error {
	resetCmd := exec.Command("git", "reset", "--hard")
	resetCmd.Dir = localDir
	if err := resetCmd.Run(); err != nil {
		return err
	}

	pullCmd := exec.Command("git", "pull", "--ff-only")
	pullCmd.Dir = localDir
	if err := pullCmd.Run(); err != nil {
		return err
	}
	return nil
}
