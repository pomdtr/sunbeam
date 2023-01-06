package app

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

//go:embed schemas/manifest.json
var manifestSchema string

type Api struct {
	Extensions    []Extension
	ExtensionRoot string
}

type RootItem struct {
	Extension string
	Script    string
	Title     string
	Subtitle  string
	With      map[string]any
}

type Extension struct {
	Title       string        `json:"title" yaml:"title"`
	Description string        `json:"description" yaml:"description"`
	Name        string        `json:"name" yaml:"name"`
	PostInstall string        `json:"postInstall" yaml:"postInstall"`
	Root        string        `json:"root" yaml:"root"`
	Preferences []ScriptInput `json:"preferences" yaml:"preferences"`

	Requirements []ExtensionRequirement `json:"requirements" yaml:"requirements"`
	RootItems    []RootItem             `json:"rootItems" yaml:"rootItems"`
	Scripts      map[string]Script      `json:"scripts" yaml:"scripts"`
}

type ExtensionRequirement struct {
	Which    string `json:"which" yaml:"which"`
	HomePage string `json:"homePage" yaml:"homePage"`
}

func (r ExtensionRequirement) Check() bool {
	if _, err := exec.LookPath(r.Which); err != nil {
		return false
	}
	return true
}

var schema *jsonschema.Schema

func init() {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", strings.NewReader(manifestSchema)); err != nil {
		panic(err)
	}

	var err error
	schema, err = compiler.Compile("schema.json")
	if err != nil {
		panic(err)
	}
}

func (api *Api) LoadExtensions(extensionRoot string) error {
	api.ExtensionRoot = extensionRoot
	entries, err := os.ReadDir(extensionRoot)
	if err != nil {
		return fmt.Errorf("failed to read extension root: %w", err)
	}

	for _, entry := range entries {
		extensionDir := path.Join(extensionRoot, entry.Name())
		if fi, err := os.Stat(extensionDir); err != nil || !fi.IsDir() {
			continue
		}

		extensionName := entry.Name()
		manifestPath := path.Join(extensionDir, "sunbeam.yml")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		extension, err := ParseManifest(extensionName, manifestPath)
		if err != nil {
			log.Println(fmt.Errorf("error parsing manifest %s: %w", manifestPath, err))
		}

		extension.Root = extensionDir
		api.Extensions = append(api.Extensions, extension)
	}

	return nil
}

func ParseManifest(extensionName string, manifestPath string) (extension Extension, err error) {
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return extension, err
	}

	var m any
	err = yaml.Unmarshal(manifestBytes, &m)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return extension, err
	}

	err = schema.Validate(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%#v", err)
		return extension, err
	}

	err = yaml.Unmarshal(manifestBytes, &extension)
	if err != nil {
		return extension, err
	}

	for key, script := range extension.Scripts {
		script.Name = key
		extension.Scripts[key] = script
	}

	for key, rootItem := range extension.RootItems {
		rootItem.Subtitle = extension.Title
		rootItem.Extension = extensionName
		extension.RootItems[key] = rootItem
	}

	extension.Name = extensionName
	extension.Root = path.Dir(manifestPath)

	return extension, nil
}
