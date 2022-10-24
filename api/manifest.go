package api

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path"

	"github.com/adrg/xdg"
)

type CommandMap map[string]Command

var (
	ExtensionMap = make(map[string]CommandMap)
	Commands     = make([]Command, 0)
)

func init() {
	manifests := listManifests()
	for _, manifest := range manifests {
		commands := manifest.Commands
		rootPath := path.Dir(manifest.Url.Path)
		rootUrl := url.URL{
			Scheme: "file",
			Path:   rootPath,
		}

		commandMap := make(map[string]Command)
		for _, command := range commands {
			command.Root = url.URL{
				Scheme: "file",
				Path:   rootPath,
			}
			command.Url = url.URL{
				Scheme: "file",
				Path:   path.Join(rootPath, command.Id),
			}
			Commands = append(Commands, command)
			commandMap[command.Id] = command
		}

		ExtensionMap[rootUrl.String()] = commandMap
	}
}

func listManifests() []Manifest {
	commandDirs := make([]string, 0)
	currentDir, err := os.Getwd()
	if err == nil {
		for currentDir != "/" {
			commandDirs = append(commandDirs, currentDir)
			currentDir = path.Dir(currentDir)
		}
	}

	var scriptDir = os.Getenv("SUNBEAM_COMMAND_DIR")
	if scriptDir == "" {
		scriptDir = xdg.ConfigHome + "/sunbeam/commands"
	}
	commandDirs = append(commandDirs, scriptDir)

	extensionRoot := path.Join(xdg.DataHome, "sunbeam", "extensions")
	extensionDirs, _ := ioutil.ReadDir(extensionRoot)
	for _, extensionDir := range extensionDirs {
		extensionPath := path.Join(extensionRoot, extensionDir.Name())
		commandDirs = append(commandDirs, extensionPath)
	}

	manifests := make([]Manifest, 0)
	for _, commandDir := range commandDirs {
		manifestPath := path.Join(commandDir, "api.json")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		var manifest Manifest
		manifestBytes, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		json.Unmarshal(manifestBytes, &manifest)
		manifest.Url = url.URL{
			Scheme: "file",
			Path:   manifestPath,
		}
		manifests = append(manifests, manifest)
	}

	return manifests
}

type Manifest struct {
	Title    string    `json:"title"`
	Commands []Command `json:"commands"`
	Url      url.URL
}
