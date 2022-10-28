package api

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"path"
)

type Api struct {
	Extensions map[string]Manifest
	RootItems  []RootItem
}

var Sunbeam Api

func init() {
	extensions, err := fetchExtensions()
	if err != nil {
		log.Fatalf("Failed to fetch manifests: %v", err)
	}

	rootItems := make([]RootItem, 0)
	for _, manifest := range extensions {
		for _, rootItem := range manifest.RootItems {
			rootItem.Extension = manifest.Name
			rootItems = append(rootItems, rootItem)
		}
	}

	Sunbeam = Api{
		Extensions: extensions,
		RootItems:  rootItems,
	}
}

func (api Api) GetCommand(extensionName string, commandName string) (SunbeamCommand, bool) {
	manifest, ok := api.Extensions[extensionName]
	if !ok {
		return SunbeamCommand{}, false
	}
	command, ok := manifest.Commands[commandName]
	if !ok {
		return SunbeamCommand{}, false
	}

	return command, true
}

var Extensions map[string]Manifest

func init() {
	var err error
	Extensions, err = fetchExtensions()
	if err != nil {
		log.Fatalf("Failed to fetch manifests: %v", err)
	}
}

type Manifest struct {
	Title string `json:"title"`
	Name  string `json:"name"`

	RootItems []RootItem                `json:"rootItems"`
	Commands  map[string]SunbeamCommand `json:"commands"`

	Url url.URL
}

type RootItem struct {
	Title     string            `json:"title"`
	Subtitle  string            `json:"subtitle"`
	Target    string            `json:"target"`
	Params    map[string]string `json:"params"`
	Extension string
}

func fetchExtensions() (map[string]Manifest, error) {
	commandDirs := make([]string, 0)
	currentDir, err := os.Getwd()
	if err == nil {
		for currentDir != "/" {
			commandDirs = append(commandDirs, currentDir)
			currentDir = path.Dir(currentDir)
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	var scriptDir = os.Getenv("SUNBEAM_COMMAND_DIR")
	if scriptDir == "" {
		scriptDir = path.Join(homeDir, ".config", "sunbeam", "commands")
	}
	commandDirs = append(commandDirs, scriptDir)

	extensionRoot := path.Join(homeDir, ".local", "share", "sunbeam", "extensions")
	extensionDirs, _ := os.ReadDir(extensionRoot)
	for _, extensionDir := range extensionDirs {
		extensionPath := path.Join(extensionRoot, extensionDir.Name())
		commandDirs = append(commandDirs, extensionPath)
	}

	manifests := make(map[string]Manifest)
	for _, commandDir := range commandDirs {
		manifestPath := path.Join(commandDir, "sunbeam.json")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		var manifest Manifest
		manifestBytes, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		err = json.Unmarshal(manifestBytes, &manifest)
		if err != nil {
			continue
		}
		manifest.Url = url.URL{
			Scheme: "file",
			Path:   manifestPath,
		}

		for key, command := range manifest.Commands {
			command.Root = url.URL{
				Scheme: "file",
				Path:   commandDir,
			}
			command.Url = url.URL{
				Scheme: "file",
				Path:   path.Join(commandDir, key),
			}
			command.Extension = manifest.Name
			manifest.Commands[key] = command
		}
		manifests[manifest.Name] = manifest
	}

	return manifests, nil
}
