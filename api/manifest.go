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
}

var Sunbeam Api

func init() {
	scriptDirs := make([]string, 0)

	currentDir, err := os.Getwd()
	if err == nil {
		for currentDir != "/" {
			scriptDirs = append(scriptDirs, currentDir)
			currentDir = path.Dir(currentDir)
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("could not get home directory: %v", err)
	}
	var scriptDir = os.Getenv("SUNBEAM_SCRIPT_DIR")
	if scriptDir == "" {
		scriptDir = path.Join(homeDir, ".config", "sunbeam", "scripts")
	}
	scriptDirs = append(scriptDirs, scriptDir)

	extensionRoot := path.Join(homeDir, ".local", "share", "sunbeam", "extensions")
	extensionDirs, _ := os.ReadDir(extensionRoot)
	for _, extensionDir := range extensionDirs {
		extensionPath := path.Join(extensionRoot, extensionDir.Name())
		scriptDirs = append(scriptDirs, extensionPath)
	}

	manifests := make(map[string]Manifest)
	for _, scriptDir := range scriptDirs {
		manifestPath := path.Join(scriptDir, "sunbeam.json")
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

		for key, script := range manifest.Scripts {
			script.Root = url.URL{
				Scheme: "file",
				Path:   scriptDir,
			}
			script.Url = url.URL{
				Scheme: "file",
				Path:   path.Join(scriptDir, key),
			}
			script.Extension = manifest.Name
			manifest.Scripts[key] = script
		}
		manifests[manifest.Name] = manifest
	}

	Sunbeam = Api{
		Extensions: manifests,
	}
}

func (api Api) GetScript(extensionName string, scriptName string) (SunbeamScript, bool) {
	manifest, ok := api.Extensions[extensionName]
	if !ok {
		return SunbeamScript{}, false
	}
	script, ok := manifest.Scripts[scriptName]
	if !ok {
		return SunbeamScript{}, false
	}

	return script, true
}

type Manifest struct {
	Title string `json:"title"`
	Name  string `json:"name"`

	RootItems []ListItem               `json:"rootItems"`
	Scripts   map[string]SunbeamScript `json:"scripts"`

	Url url.URL
}
