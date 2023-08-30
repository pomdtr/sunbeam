package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GithubRepo struct {
	Name            string
	FullName        string `json:"full_name"`
	Description     string
	HtmlUrl         string `json:"html_url"`
	StargazersCount int    `json:"stargazers_count"`
	Topics          []string
}

func FetchGithubRepository(owner string, name string) (*GithubRepo, error) {
	res, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, name))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get repo metadata: %s", res.Status)
	}

	var repo GithubRepo
	if err := json.NewDecoder(res.Body).Decode(&repo); err != nil {
		return nil, err
	}

	return &repo, nil
}

type GithubGist struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
}

func FetchGithubGist(gistID string) (*GithubGist, error) {
	res, err := http.Get(fmt.Sprintf("https://api.github.com/gists/%s", gistID))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get gist metadata: %s", res.Status)
	}

	var gist GithubGist
	if err := json.NewDecoder(res.Body).Decode(&gist); err != nil {
		return nil, err
	}

	return &gist, nil
}

type GitCommit struct {
	Sha string `json:"sha"`
}

func GetLastGitCommit(owner string, name string) (*GitCommit, error) {
	apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits", owner, name)

	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get latest release: %s", resp.Status)
	}

	var commits []*GitCommit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found")
	}

	return commits[0], nil
}

type GistCommit struct {
	Version string `json:"version"`
}

func GetLastGistCommit(gistID string) (*GistCommit, error) {
	res, err := http.Get(fmt.Sprintf("https://api.github.com/gists/%s/commits", gistID))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get gist metadata: %s", res.Status)
	}

	var commits []*GistCommit
	if err := json.NewDecoder(res.Body).Decode(&commits); err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found")
	}

	return commits[0], nil
}
