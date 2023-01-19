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
	Extensions    map[string]Extension
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
	Extension string         `json:"extension,omitempty"`
	Command   string         `json:"command"`
	Title     string         `json:"title"`
	With      map[string]any `json:"with,omitempty"`
}

type Extension struct {
	Title       string      `json:"title" yaml:"title"`
	Description string      `json:"description,omitempty" yaml:"description"`
	PostInstall string      `json:"postInstall,omitempty" yaml:"postInstall"`
	Root        string      `json:"root,omitempty" yaml:"root"`
	Preferences []FormInput `json:"preferences,omitempty"`

	Requirements []ExtensionRequirement `json:"requirements,omitempty"`
	RootItems    []RootItem             `json:"rootItems" yaml:"rootItems"`
	Commands     map[string]Command     `json:"commands"`
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
	api.Extensions = make(map[string]Extension)
	entries, err := os.ReadDir(extensionRoot)
	if err != nil {
		return fmt.Errorf("failed to read extension root: %w", err)
	}

	for _, entry := range entries {
		extensionDir := path.Join(extensionRoot, entry.Name())
		if fi, err := os.Stat(extensionDir); err != nil || !fi.IsDir() {
			continue
		}

		manifestPath := path.Join(extensionDir, "sunbeam.yml")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		extension, err := ParseManifest(manifestPath)
		if err != nil {
			continue
		}

		extension.Root = extensionDir
		api.Extensions[entry.Name()] = extension
	}

	return nil
}

func ParseManifest(manifestPath string) (extension Extension, err error) {
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

	extension.Root = path.Dir(manifestPath)
	return extension, nil
}
