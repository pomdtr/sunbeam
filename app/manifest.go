package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/SunbeamLauncher/sunbeam/utils"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

//go:embed schemas/manifest.json
var manifestSchema string

type Api struct {
	Extensions       map[string]Extension
	ExtensionRoot    string
	ExtensionConfigs map[string]ExtensionConfig
	ConfigRoot       string
}

var Sunbeam Api

func (api Api) AddExtension(name string, config ExtensionConfig) error {
	if _, ok := api.ExtensionConfigs[name]; ok {
		return fmt.Errorf("extension %s already exists", name)
	}
	api.ExtensionConfigs[name] = config
	bytes, err := json.Marshal(api.ExtensionConfigs)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(api.ExtensionRoot, "extensions.json"), bytes, 0644)
}

func (api Api) RemoveExtension(configName string) error {
	config, ok := api.ExtensionConfigs[configName]
	if !ok {
		return fmt.Errorf("extension %s does not exist", configName)
	}

	if err := os.RemoveAll(config.Root); err != nil {
		return err
	}

	delete(api.ExtensionConfigs, configName)
	bytes, err := json.Marshal(api.ExtensionConfigs)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(api.ExtensionRoot, "extensions.json"), bytes, 0644)
}

type RootItem struct {
	Extension string
	Script    string
	Title     string
	Subtitle  string
	With      map[string]any
}

type ExtensionConfig struct {
	Root   string `json:"root"`
	Remote string `json:"remote,omitempty"`
}

type ExtensionManifest []ExtensionConfig

type Extension struct {
	Title       string        `json:"title" yaml:"title"`
	Description string        `json:"description" yaml:"description"`
	Name        string        `json:"name" yaml:"name"`
	PostInstall string        `json:"postInstall" yaml:"postInstall"`
	Preferences []ScriptParam `json:"preferences" yaml:"preferences"`

	Requirements []ExtensionRequirement `json:"requirements" yaml:"requirements"`
	RootItems    []RootItem             `json:"rootItems" yaml:"rootItems"`
	Scripts      map[string]Script      `json:"scripts" yaml:"scripts"`

	Url url.URL
}

type ExtensionRequirement struct {
	Which    string `json:"which" yaml:"which"`
	HomePage string `json:"homePage" yaml:"homePage"`
}

func (r ExtensionRequirement) Check() bool {
	cmd := exec.Command("which", r.Which)
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func (m Extension) Dir() string {
	return path.Dir(m.Url.Path)
}

func init() {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", strings.NewReader(manifestSchema)); err != nil {
		panic(err)
	}
	schema, err := compiler.Compile("schema.json")
	if err != nil {
		panic(err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("could not get home directory: %v", err)
	}

	extensionRoot := path.Join(homeDir, ".local", "share", "sunbeam", "extensions")
	if _, err := os.Stat(extensionRoot); os.IsNotExist(err) {
		os.MkdirAll(extensionRoot, 0755)
	}

	extensionManifestPath := path.Join(extensionRoot, "extensions.json")
	if _, err := os.Stat(extensionManifestPath); os.IsNotExist(err) {
		os.WriteFile(extensionManifestPath, []byte("{}"), 0644)
	}

	var ExtensionConfigs map[string]ExtensionConfig
	if err := utils.ReadJson(extensionManifestPath, &ExtensionConfigs); err != nil {
		log.Fatalf("could not read extension manifest: %v", err)
	}

	extensions := make(map[string]Extension)
	for _, extensionConfig := range ExtensionConfigs {
		manifestPath := path.Join(extensionConfig.Root, "sunbeam.yml")
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
		ExtensionRoot:    extensionRoot,
		ExtensionConfigs: ExtensionConfigs,
		ConfigRoot:       path.Join(homeDir, ".config", "sunbeam"),
		Extensions:       extensions,
	}
}

func ParseManifest(bytes []byte) (extension Extension, err error) {
	err = yaml.Unmarshal(bytes, &extension)
	if err != nil {
		return extension, err

	}

	for key, script := range extension.Scripts {
		script.Name = key
		extension.Scripts[key] = script
	}

	for key, rootItem := range extension.RootItems {
		rootItem.Subtitle = extension.Title
		rootItem.Extension = extension.Name
		extension.RootItems[key] = rootItem
	}

	return extension, nil
}
