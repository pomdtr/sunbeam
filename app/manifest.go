package app

import (
	_ "embed"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

//go:embed manifest.json
var manifestSchema string

type Api struct {
	Extensions    map[string]Extension
	ExtensionRoot string
	ConfigRoot    string
}

type RootItem struct {
	Extension string
	Script    string
	Title     string
	Subtitle  string
	With      ScriptInputs
}

type Extension struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	Name        string `json:"name" yaml:"name"`
	PostInstall string `json:"postInstall" yaml:"postInstall"`

	RootItems []RootItem        `json:"rootItems" yaml:"rootItems"`
	Scripts   map[string]Script `json:"scripts" yaml:"scripts"`

	Url url.URL
}

func (m Extension) Dir() string {
	return path.Dir(m.Url.Path)
}

var Sunbeam Api

func init() {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", strings.NewReader(manifestSchema)); err != nil {
		panic(err)
	}
	schema, err := compiler.Compile("schema.json")
	if err != nil {
		panic(err)
	}
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

		var m any
		err = yaml.Unmarshal(manifestBytes, &m)
		if err != nil {
			panic(err)
		}

		err = schema.Validate(m)
		if err != nil {
			log.Println(fmt.Errorf("error validating manifest %s: %w", manifestPath, err))
			continue
		}

		extension, err := ParseManifest(manifestBytes)
		if err != nil {
			log.Println(fmt.Errorf("error parsing manifest %s: %w", manifestPath, err))
		}

		extension.Url = url.URL{
			Scheme: "file",
			Path:   manifestPath,
		}

		extensions[extension.Name] = extension
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
