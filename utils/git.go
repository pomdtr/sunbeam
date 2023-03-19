package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
)

type Repository struct {
	url *url.URL
}

func NewRepository(url *url.URL) *Repository {
	return &Repository{url: url}
}

func (r *Repository) FullName() string {
	return fmt.Sprintf("%s/%s", r.Owner(), r.Name())
}

func (r *Repository) Owner() string {
	return path.Base(path.Dir(r.url.Path))
}

func (r *Repository) Name() string {
	return path.Base(r.url.Path)
}

func RepositoryFromString(repositoryUrl string) (*Repository, error) {
	parsedUrl, err := url.Parse(repositoryUrl)
	if err != nil {
		return nil, err
	}

	return NewRepository(parsedUrl), nil
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
