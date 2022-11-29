package app

import (
	"log"
	"net/url"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type Api struct {
	Extensions    map[string]Extension
	ExtensionRoot string
	ConfigRoot    string
}

type Entrypoint struct {
	Script string
	Title  string
	With   ScriptArguments
}

type Extension struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	Id          string `json:"id" yaml:"id"`
	PostInstall string `json:"postInstall" yaml:"postInstall"`

	RootItems []Entrypoint      `json:"rootItems" yaml:"rootItems"`
	Scripts   map[string]Script `json:"scripts" yaml:"scripts"`

	Url url.URL
}

func (m Extension) Dir() string {
	return path.Dir(m.Url.Path)
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
	extensionRoot := path.Join(homeDir, ".local", "share", "sunbeam", "extensions")
	configRoot := path.Join(homeDir, ".config", "sunbeam")
	extensionDirs, _ := os.ReadDir(extensionRoot)
	for _, extensionDir := range extensionDirs {
		extensionPath := path.Join(extensionRoot, extensionDir.Name())
		scriptDirs = append(scriptDirs, extensionPath)
	}

	extensions := make(map[string]Extension)
	for _, scriptDir := range scriptDirs {

		manifestPath := path.Join(scriptDir, "sunbeam.yml")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		manifestBytes, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		extension, err := ParseManifest(manifestBytes)
		if err != nil {
			log.Println(err)
		}

		extension.Url = url.URL{
			Scheme: "file",
			Path:   manifestPath,
		}

		extensions[extension.Id] = extension
	}

	Sunbeam = Api{
		ExtensionRoot: extensionRoot,
		ConfigRoot:    configRoot,
		Extensions:    extensions,
	}
}

func ParseManifest(bytes []byte) (extension Extension, err error) {
	err = yaml.Unmarshal(bytes, &extension)
	if err != nil {
		return extension, err

	}

	for key, script := range extension.Scripts {
		if script.Page.Title == "" {
			script.Page.Title = extension.Title
		}
		extension.Scripts[key] = script
	}

	return extension, nil
}
