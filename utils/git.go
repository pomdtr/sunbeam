package utils

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

type GitClient struct {
	repo string
}

func NewGitClient(repo string) *GitClient {
	return &GitClient{
		repo: repo,
	}
}

func GitClone(url string, target string) error {
	cmd := exec.Command("git", "clone", "--filter=blob:none", url, target)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func (gc *GitClient) Pull() error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = gc.repo
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func (gc *GitClient) Config(name string) (string, error) {
	cmd := exec.Command("git", "config", name)
	cmd.Dir = gc.repo
	cmd.Stderr = os.Stderr
	res, error := cmd.Output()

	return string(res), error
}

func (gc *GitClient) GetOrigin() string {
	origin, _ := gc.Config("remote.origin.url")
	return strings.TrimSpace(origin)
}

// getCurrentVersion determines the current version for non-local git extensions.
func (gc *GitClient) GetCurrentVersion() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = gc.repo
	cmd.Stderr = os.Stderr
	res, err := cmd.Output()

	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(res))
}

func (gc *GitClient) GetLatestVersion() (string, error) {
	cmd := exec.Command("git", "ls-remote", "origin", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	remoteSha := bytes.SplitN(output, []byte("\t"), 2)[0]
	return string(remoteSha), nil
}
