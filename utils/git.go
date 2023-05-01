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
	url *url.URL
}

func RepositoryUrl(pattern string) (*url.URL, error) {
	return url.Parse(pattern)
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

type GithubGist struct {
	Url         string `json:"url"`
	Description string `json:"description"`
	Files       map[string]struct {
		Filename string `json:"filename"`
		Content  string `json:"content"`
	} `json:"files"`
}

func FetchGist(id string) (*GithubGist, error) {
	res, err := http.Get(fmt.Sprintf("https://api.github.com/gists/%s", id))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get gist: %s", res.Status)
	}

	var gist GithubGist
	if err := json.NewDecoder(res.Body).Decode(&gist); err != nil {
		return nil, err
	}

	return &gist, nil
}
