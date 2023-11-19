package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func LoadToken() (string, error) {
	if token, ok := os.LookupEnv("SUNBEAM_GITHUB_TOKEN"); ok {
		return token, nil
	}

	if token, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		return token, nil
	}

	return "", fmt.Errorf("github token not found in environment")
}

type CreateGistRequestBody struct {
	Description string                               `json:"description"`
	Public      bool                                 `json:"public"`
	Files       map[string]CreateGistRequestBodyFile `json:"files"`
}

type CreateGistRequestBodyFile struct {
	Content string `json:"content"`
}

type Gist struct {
	ID      string `json:"id"`
	HtmlURL string `json:"html_url"`
	Owner   struct {
		Login string `json:"login"`
	} `json:"owner"`
}

func CreateGist(filename string, content []byte, description string, public bool) (Gist, error) {
	files := make(map[string]CreateGistRequestBodyFile)
	files[filename] = CreateGistRequestBodyFile{
		Content: string(content),
	}

	body, err := json.Marshal(CreateGistRequestBody{
		Description: description,
		Public:      public,
		Files:       files,
	})
	if err != nil {
		return Gist{}, err
	}

	req, err := http.NewRequest("POST", "https://api.github.com/gists", bytes.NewReader(body))
	if err != nil {
		return Gist{}, err
	}

	token, err := LoadToken()
	if err != nil {
		return Gist{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Gist{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return Gist{}, fmt.Errorf("failed to create gist: %s", resp.Status)
	}

	var gist Gist
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return Gist{}, err
	}

	return gist, nil
}
