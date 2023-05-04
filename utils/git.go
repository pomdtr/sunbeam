package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

type Repository struct {
	URL *url.URL
}

func RepositoryUrl(pattern string) (*url.URL, error) {
	return url.Parse(pattern)
}

func NewRepository(url *url.URL) *Repository {
	return &Repository{URL: url}
}

func (r *Repository) String() string {
	return r.URL.String()
}

func (r *Repository) FullName() string {
	return fmt.Sprintf("%s/%s", r.Owner(), r.Name())
}

func (r *Repository) Owner() string {
	return filepath.Base(filepath.Dir(r.URL.Path))
}

func (r *Repository) Name() string {
	return filepath.Base(r.URL.Path)
}

type Release struct {
	TagName string `json:"tag_name"`
}

func GetLatestRelease(repo *Repository) (*Release, error) {
	if repo.URL.Host != "github.com" {
		return nil, fmt.Errorf("unsupported host: %s", repo.URL.Host)
	}

	apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo.FullName())

	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get latest release: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func GitClone(repo *Repository, targetDir string) error {
	command := exec.Command("git", "clone", repo.URL.String(), targetDir)
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
