package app

import (
	"embed"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"

	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"gopkg.in/yaml.v3"
)

//go:embed schemas
var embedFs embed.FS

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
	Extension string `json:",omitempty" yaml:",omitempty"`
	Command   string
	Title     string
	With      map[string]CommandInput `json:",omitempty" yaml:",omitempty"`
}

type Extension struct {
	Version     string   `json:"version" yaml:"version"`
	Title       string   `json:"title" yaml:"title"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	PostInstall string   `json:"postInstall,omitempty" yaml:"postInstall,omitempty"`
	RootUrl     string   `json:"rootUrl,omitempty" yaml:"rootUrl,omitempty"`
	Root        *url.URL `json:"-" yaml:"-"`
	// Preferences []Preference `json:"preferences,omitempty"`

	Requirements []ExtensionRequirement `json:"requirements,omitempty" yaml:"requirements,omitempty"`
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

var ExtensionSchema *jsonschema.Schema
var PageSchema *jsonschema.Schema

func init() {
	var err error

	compiler := jsonschema.NewCompiler()

	manifest, err := embedFs.Open("schemas/extension.json")
	if err != nil {
		panic(err)
	}
	if err = compiler.AddResource("https://pomdtr.github.io/sunbeam/schemas/extension.json", manifest); err != nil {
		panic(err)
	}

	page, err := embedFs.Open("schemas/page.json")
	if err != nil {
		panic(err)
	}
	if err = compiler.AddResource("https://pomdtr.github.io/sunbeam/schemas/page.json", page); err != nil {
		panic(err)
	}

	ExtensionSchema, err = compiler.Compile("https://pomdtr.github.io/sunbeam/schemas/extension.json")
	if err != nil {
		panic(err)
	}

	PageSchema, err = compiler.Compile("https://pomdtr.github.io/sunbeam/schemas/page.json")
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

		if extension.RootUrl != "" {
			root, err := url.Parse(extension.RootUrl)
			if err != nil {
				continue
			}
			extension.Root = root
		} else {
			extension.Root = &url.URL{
				Scheme: "file",
				Path:   extensionDir,
			}
		}

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

	err = ExtensionSchema.Validate(m)
	if err != nil {
		return extension, err
	}

	err = yaml.Unmarshal(manifestBytes, &extension)
	if err != nil {
		return extension, err
	}

	return extension, nil
}
