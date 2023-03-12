package utils

import (
	"fmt"
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

func (r *Repository) Fullname() string {
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
