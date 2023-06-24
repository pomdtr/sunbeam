package store

import (
	"encoding/json"
	"net/http"
)

type ItemType string

const (
	GithubRelease ItemType = "github-release"
	GithubRepo    ItemType = "github-repository"
	GithubGist    ItemType = "github-gist"
	Script        ItemType = "script"
	API           ItemType = "api"
)

type Platform string

const (
	Windows Platform = "windows"
	Linux   Platform = "linux"
	MacOS   Platform = "macos"
)

type CatalogItem struct {
	Type        ItemType
	Title       string
	Description string
	Author      string
	Origin      string
	Platforms   []Platform
}

type Catalog struct {
	Extensions []CatalogItem
}

func FetchCatalog() ([]CatalogItem, error) {
	resp, err := http.Get("https://raw.githubusercontent.com/pomdtr/sunbeam/main/store/catalog.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var catalog Catalog
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return nil, err
	}

	return catalog.Extensions, nil
}
