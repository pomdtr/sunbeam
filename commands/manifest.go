package commands

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"path"

	"github.com/adrg/xdg"
)

var Commands map[string]Command = make(map[string]Command)
var RootItems []ListItem

func init() {
	var scriptDir = os.Getenv("SUNBEAM_SCRIPT_DIR")
	if scriptDir == "" {
		scriptDir = xdg.DataHome
	}

	manifestJson := path.Join(scriptDir, "sunbeam.json")
	manifest, err := ParseManifest(manifestJson)
	if err != nil {
		log.Fatalf("error while parsing manifest: %s", err)
	}

	for _, commandData := range manifest.Commands {
		command := Command{}
		command.CommandData = commandData
		command.RootUrl = url.URL{
			Scheme: "file",
			Path:   scriptDir,
		}
		command.Url = url.URL{
			Scheme: "file",
			Path:   path.Join(scriptDir, command.Command),
		}
		Commands[command.Id] = command
	}

	RootItems = make([]ListItem, 0)
	for commandId, command := range Commands {
		RootItems = append(RootItems, ListItem{
			Title:    command.Title,
			Subtitle: command.Subtitle,
			Actions: []ScriptAction{
				{
					Type:   "push",
					Title:  "Open Command",
					Target: commandId,
				},
			},
		})
	}
}

type Manifest struct {
	Title    string        `json:"title"`
	Commands []CommandData `json:"commands"`
}

func ParseManifest(manifestPath string) (*Manifest, error) {
	var manifest Manifest
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(manifestBytes, &manifest)

	return &manifest, nil
}
