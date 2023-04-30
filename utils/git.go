package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

type Repository struct {
	url *url.URL
}

func NewRepository(url *url.URL) *Repository {
	return &Repository{url: url}
}

func (r *Repository) String() string {
	return r.url.String()
}

func (r *Repository) FullName() string {
	return fmt.Sprintf("%s/%s", r.Owner(), r.Name())
}

func (r *Repository) Owner() string {
	return filepath.Base(filepath.Dir(r.url.Path))
}

func (r *Repository) Name() string {
	return filepath.Base(r.url.Path)
}

var ownerNameRegexp = regexp.MustCompile(`^([a-zA-Z0-9-]+)\/([a-zA-Z0-9-]+)$`)
var urlRegexp = regexp.MustCompile(`^https?://github\.com/[a-zA-Z0-9][a-zA-Z0-9-]+/[a-zA-Z0-9_.-]+$`)

func RepositoryFromString(pattern string) (*Repository, error) {
	if ownerNameRegexp.MatchString(pattern) {
		return NewRepository(&url.URL{
			Scheme: "https",
			Host:   "github.com",
			Path:   pattern,
		}), nil
	}

	if urlRegexp.MatchString(pattern) {
		repoUrl, err := url.Parse(pattern)
		if err != nil {
			return nil, err
		}

		return NewRepository(repoUrl), nil
	}

	return nil, fmt.Errorf("invalid repository pattern: %s", pattern)
}

func (r *Repository) Url() *url.URL {
	return r.url
}

type Release struct {
	TagName string `json:"tag_name"`
}

func GetLatestRelease(repo *Repository) (*Release, error) {
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
	command := exec.Command("git", "clone", repo.Url().String(), targetDir)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}

func GitPull(localDir string) error {
	command := exec.Command("git", "pull")
	command.Dir = localDir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}
