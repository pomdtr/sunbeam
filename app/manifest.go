package app

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

//go:embed schemas/manifest.json
var manifest string

type Api struct {
	Extensions    []Extension
	ExtensionRoot string
}

func (api *Api) IsExtensionInstalled(name string) bool {
	extensionDir := path.Join(api.ExtensionRoot, name)

	if _, err := os.Stat(extensionDir); os.IsNotExist(err) {
		return false
	}
	return true
}

type RootItem struct {
	Extension string
	Command   string
	Title     string
	Subtitle  string
	With      map[string]any
}

type Extension struct {
	Title       string
	Description string
	Name        string
	PostInstall string `json:"postInstall" yaml:"postInstall"`
	Root        string
	Preferences []ScriptInput

	Requirements []ExtensionRequirement
	RootItems    []RootItem `json:"rootItems" yaml:"rootItems"`
	Commands     map[string]Command
}

type ExtensionRequirement struct {
	Which    string
	HomePage string `json:"homePage" yaml:"homePage"`
}

func (r ExtensionRequirement) Check() bool {
	if _, err := exec.LookPath(r.Which); err != nil {
		return false
	}
	return true
}

var ManifestSchema *jsonschema.Schema

func init() {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("manifest", strings.NewReader(manifest)); err != nil {
		panic(err)
	}

	var err error
	ManifestSchema, err = compiler.Compile("manifest")
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
			continue
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

	err = ManifestSchema.Validate(m)
	if err != nil {
		return extension, err
	}

	err = yaml.Unmarshal(manifestBytes, &extension)
	if err != nil {
		return extension, err
	}

	for key, command := range extension.Commands {
		command.Name = key
		extension.Commands[key] = command
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
