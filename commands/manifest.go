package commands

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
	commandRoots := listCommandDirs()
	for _, commandRoot := range commandRoots {
		dirCommands, err := listCommands(commandRoot)
		if err != nil {
			log.Printf("Error while fetching commands: %s", err)
			continue
		}

		commandMap := make(map[string]Command)
		for _, command := range dirCommands {
			Commands = append(Commands, command)
			commandMap[command.Id] = command
		}

		ExtensionMap[commandRoot.String()] = commandMap
	}
}

func listCommands(commandRoot url.URL) (commands []Command, err error) {
	manifestPath := path.Join(commandRoot.Path, "sunbeam.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, err
	}

	var manifest Manifest
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(manifestBytes, &manifest)

	if err != nil {
		log.Fatalf("error while parsing manifest: %s", err)
	}

	for _, commandData := range manifest.Commands {
		command := Command{}
		command.CommandData = commandData
		rootPath := path.Dir(manifestPath)
		command.Root = url.URL{
			Scheme: "file",
			Path:   rootPath,
		}
		command.Url = url.URL{
			Scheme: "file",
			Path:   path.Join(rootPath, command.Command),
		}
		commands = append(commands, command)
	}

	return commands, nil
}

func listCommandDirs() (commandDirs []url.URL) {
	currentDir, err := os.Getwd()
	if err == nil {
		commandDirs = append(commandDirs, url.URL{
			Scheme: "file",
			Path:   currentDir,
		})
	}

	var scriptDir = os.Getenv("SUNBEAM_COMMAND_DIR")
	if scriptDir == "" {
		scriptDir = xdg.ConfigHome + "/sunbeam/commands"
	}
	commandDirs = append(commandDirs, url.URL{
		Scheme: "file",
		Path:   scriptDir,
	})

	extensionRoot := path.Join(xdg.DataHome, "sunbeam", "extensions")
	extensionDirs, _ := ioutil.ReadDir(extensionRoot)
	for _, extensionDir := range extensionDirs {
		extensionPath := path.Join(extensionRoot, extensionDir.Name())
		commandDirs = append(commandDirs, url.URL{
			Scheme: "file",
			Path:   extensionPath,
		})
	}

	return commandDirs
}

type Manifest struct {
	Title    string        `json:"title"`
	Commands []CommandData `json:"commands"`
}
