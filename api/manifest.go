package api

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
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
		if _, ok := ExtensionMap[manifest.Id]; ok {
			continue
		}

		commandMap := make(map[string]Command)
		for _, command := range commands {
			command.Root = url.URL{
				Scheme: "file",
				Path:   rootPath,
			}
			command.ExtensionId = manifest.Id
			command.Url = url.URL{
				Scheme: "file",
				Path:   path.Join(rootPath, command.Id),
			}
			Commands = append(Commands, command)
			commandMap[command.Id] = command
		}

		ExtensionMap[manifest.Id] = commandMap
	}
}

func GetCommand(target string) (Command, bool) {
	tokens := strings.Split(target, "/")
	if len(tokens) < 2 {
		return Command{}, false
	}

	extension, ok := ExtensionMap[tokens[0]]
	if !ok {
		return Command{}, false
	}

	command, ok := extension[tokens[1]]
	if !ok {
		return Command{}, false
	}

	return command, true
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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []Manifest{}
	}
	var scriptDir = os.Getenv("SUNBEAM_COMMAND_DIR")
	if scriptDir == "" {
		scriptDir = path.Join(homeDir, ".config", "sunbeam", "commands")
	}
	commandDirs = append(commandDirs, scriptDir)

	extensionRoot := path.Join(homeDir, ".local", "share", "sunbeam", "extensions")
	extensionDirs, _ := ioutil.ReadDir(extensionRoot)
	for _, extensionDir := range extensionDirs {
		extensionPath := path.Join(extensionRoot, extensionDir.Name())
		commandDirs = append(commandDirs, extensionPath)
	}

	manifests := make([]Manifest, 0)
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
	Id       string    `json:"id"`
	Commands []Command `json:"commands"`
	Url      url.URL
}
