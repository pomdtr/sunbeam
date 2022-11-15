package api

import (
	"bytes"
	"io"
	"log"
	"net/url"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type Api struct {
	Extensions    map[string]Manifest
	ExtensionRoot string
}

type Manifest struct {
	Title string `json:"title" yaml:"title"`
	Name  string `json:"name" yaml:"name"`

	Entrypoints []Action          `json:"entrypoints" yaml:"entrypoints"`
	Pages       map[string]Script `json:"pages" yaml:"pages"`

	Url url.URL
}

func (m Manifest) Dir() string {
	return path.Dir(m.Url.Path)
}

var Sunbeam Api

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("could not get home directory: %v", err)
	}
	Sunbeam.ExtensionRoot = path.Join(homeDir, ".local", "share", "sunbeam", "extensions")

	scriptDirs := make([]string, 0)

	currentDir, err := os.Getwd()
	if err == nil {
		for currentDir != "/" {
			scriptDirs = append(scriptDirs, currentDir)
			currentDir = path.Dir(currentDir)
		}
	}

	var scriptDir = os.Getenv("SUNBEAM_SCRIPT_DIR")
	if scriptDir == "" {
		scriptDir = path.Join(homeDir, ".config", "sunbeam")
	}
	scriptDirs = append(scriptDirs, scriptDir)

	extensionDirs, _ := os.ReadDir(Sunbeam.ExtensionRoot)
	for _, extensionDir := range extensionDirs {
		extensionPath := path.Join(Sunbeam.ExtensionRoot, extensionDir.Name())
		scriptDirs = append(scriptDirs, extensionPath)
	}

	manifests := make(map[string]Manifest)
	for _, scriptDir := range scriptDirs {
		manifestPath := path.Join(scriptDir, "sunbeam.yml")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		manifestBytes, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		decoder := yaml.NewDecoder(bytes.NewReader(manifestBytes))
		var manifest Manifest

		for {
			err := decoder.Decode(&manifest)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("error decoding manifest at %s: %v", manifestPath, err)
				continue
			}

			manifest.Url = url.URL{
				Scheme: "file",
				Path:   manifestPath,
			}

			manifests[manifest.Name] = manifest
		}
	}

	Sunbeam = Api{
		Extensions: manifests,
	}
}
