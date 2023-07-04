package catalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
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

type CommandMetadata struct {
	SchemaVersion int         `json:"schemaVersion"`
	Title         string      `json:"title"`
	Description   string      `json:"description,omitempty"`
	Mode          CommandMode `json:"mode"`
	Author        string      `json:"author,omitempty"`
	AuthorURL     string      `json:"authorURL,omitempty"`
}

type CommandMode string

const (
	CommandModeFullOutput CommandMode = "fullOutput"
	CommandModeSilent     CommandMode = "silent"
	CommandModeView       CommandMode = "view"
)

func (m CommandMode) Supported() bool {
	return m == CommandModeFullOutput || m == CommandModeSilent || m == CommandModeView
}

var MetadataRegexp = regexp.MustCompile(`@(sunbeam|raycast)\.(?P<key>[A-Za-z0-9]+)\s(?P<value>[\S ]+)`)

func ExtractCommandMetadata(script []byte) (*CommandMetadata, error) {
	matches := MetadataRegexp.FindAllSubmatch(script, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no metadata found")
	}

	metadata := CommandMetadata{}
	for _, match := range matches {
		key := string(match[2])
		value := string(match[3])

		switch key {
		case "title":
			metadata.Title = value
		case "mode":
			metadata.Mode = CommandMode(value)
			if !metadata.Mode.Supported() {
				return nil, fmt.Errorf("unsupported mode: %s", metadata.Mode)
			}
		case "schemaVersion":
			if err := json.Unmarshal([]byte(value), &metadata.SchemaVersion); err != nil {
				return nil, fmt.Errorf("unable to parse schemaVersion: %s", err)
			}
		default:
			log.Printf("unsupported metadata key: %s", key)
		}
	}

	if metadata.SchemaVersion == 0 {
		return nil, fmt.Errorf("no schemaVersion found")
	}

	if metadata.Title == "" {
		return nil, fmt.Errorf("no title found")
	}

	if metadata.Mode == "" {
		return nil, fmt.Errorf("no mode found")
	}

	return &metadata, nil
}

type CatalogItem struct {
	Metadata *CommandMetadata
	Origin   string `json:"origin"`
}

func FetchCatalog() ([]CatalogItem, error) {
	resp, err := http.Get("https://pomdtr.github.io/sunbeam/static/catalog.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var catalog []CatalogItem
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return nil, err
	}

	return catalog, nil
}
