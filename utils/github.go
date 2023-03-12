package utils

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type GithubRepo struct {
	Name            string
	FullName        string `json:"full_name"`
	Description     string
	HtmlUrl         string `json:"html_url"`
	StargazersCount int    `json:"stargazers_count"`
}

type SearchBody struct {
	Items []GithubRepo
}

func SearchSunbeamExtensions(query string) ([]GithubRepo, error) {

	// Search extension with a sunbeam-extension topic
	extensionUrl := url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   "/search/repositories",
		RawQuery: url.Values{
			"q":     []string{"topic:sunbeam-extension"},
			"sort":  []string{"stars"},
			"order": []string{"desc"},
		}.Encode(),
	}

	res, err := http.Get(extensionUrl.String())
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, err
	}

	var body SearchBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Items, nil
}
