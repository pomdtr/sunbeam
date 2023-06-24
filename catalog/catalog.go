package catalog

import (
	_ "embed"
	"encoding/json"
	"net/http"
)

//go:embed catalog.json
var catalogRaw []byte

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

func GetCatalog() ([]CatalogItem, error) {
	var catalog Catalog
	if err := json.Unmarshal(catalogRaw, &catalog); err != nil {
		return nil, err
	}

	return catalog.Extensions, nil
}

func FetchCatalog() ([]CatalogItem, error) {
	resp, err := http.Get("https://raw.githubusercontent.com/pomdtr/sunbeam/main/catalog/catalog.json")
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
