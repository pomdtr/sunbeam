package store

type ItemType string

const (
	GithubRelease ItemType = "github-release"
	GitRepo       ItemType = "git"
	Gist          ItemType = "gist"
	Script        ItemType = "script"
	API           ItemType = "api"
)

type CatalogItem struct {
	Type        ItemType
	Title       string
	Description string
	Author      string
	Origin      string
}

type Catalog struct {
	Extensions []CatalogItem
}
