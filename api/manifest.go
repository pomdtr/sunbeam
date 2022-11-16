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

type Entrypoint struct {
	Script string
	Title  string
	With   map[string]any
}

type Manifest struct {
	Title string `json:"title" yaml:"title"`
	Name  string `json:"name" yaml:"name"`

	Entrypoints []Entrypoint      `json:"entrypoints" yaml:"entrypoints"`
	Scripts     map[string]Script `json:"scripts" yaml:"scripts"`

	Url url.URL
}

func (m Manifest) Dir() string {
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
	var scriptDir = os.Getenv("SUNBEAM_SCRIPT_DIR")
	if scriptDir == "" {
		scriptDir = path.Join(homeDir, ".config", "sunbeam")
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
		ExtensionRoot: extensionRoot,
		Extensions:    manifests,
	}
}
