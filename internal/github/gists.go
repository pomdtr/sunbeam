package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type GistClient struct {
	token string
}

func NewGistClient(token string) GistClient {
	return GistClient{
		token: token,
	}
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

func (g GistClient) CreateGist(filename string, content []byte, public bool) (Gist, error) {
	files := make(map[string]CreateGistRequestBodyFile)
	files[filename] = CreateGistRequestBodyFile{
		Content: string(content),
	}

	body, err := json.Marshal(CreateGistRequestBody{
		Public: public,
		Files:  files,
	})
	if err != nil {
		return Gist{}, err
	}

	req, err := http.NewRequest("POST", "https://api.github.com/gists", bytes.NewReader(body))
	if err != nil {
		return Gist{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.token))

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

func (g GistClient) PatchGistDescription(gistID string, description string) error {
	body, err := json.Marshal(map[string]string{
		"description": description,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("https://api.github.com/gists/%s", gistID), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to patch gist: %s", resp.Status)
	}

	return nil
}
