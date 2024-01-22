package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pomdtr/sunbeam/internal/schemas"
	"github.com/pomdtr/sunbeam/internal/utils"
)

var Path string

func init() {
	if env, ok := os.LookupEnv("SUNBEAM_CONFIG"); ok {
		Path = env
		return
	}

	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for currentDir != "/" {
		if _, err := os.Stat(filepath.Join(currentDir, "sunbeam.json")); err == nil {
			Path = filepath.Join(currentDir, "sunbeam.json")
			return
		}
		currentDir = filepath.Dir(currentDir)
	}

	Path = filepath.Join(utils.ConfigDir(), "sunbeam.json")
}

type Config struct {
	Oneliners  []Oneliner                 `json:"oneliners,omitempty"`
	Extensions map[string]ExtensionConfig `json:"extensions,omitempty"`
	path       string                     `json:"-"`
}

func (cfg Config) Resolve(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(os.Getenv("HOME"), path[2:])
	}

	if !filepath.IsAbs(path) {
		return filepath.Join(filepath.Dir(cfg.path), path)
	}

	return path
}

type ExtensionConfig struct {
	Origin      string         `json:"origin,omitempty"`
	Preferences map[string]any `json:"preferences,omitempty"`
	Root        []RootItem     `json:"root,omitempty"`
}

type RootItem struct {
	Title   string         `json:"title"`
	Command string         `json:"command"`
	Params  map[string]any `json:"params,omitempty"`
}

type Oneliner struct {
	Title       string `json:"title"`
	Command     string `json:"command"`
	Interactive bool   `json:"interactive,omitempty"`
	Cwd         string `json:"cwd,omitempty"`
	Exit        bool   `json:"exit,omitempty"`
}

func (cfg Config) Aliases() []string {
	var aliases []string
	for alias := range cfg.Extensions {
		aliases = append(aliases, alias)
	}

	return aliases
}

func Load(configPath string) (Config, error) {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load config: %w", err)
	}

	if err := schemas.ValidateConfig(configBytes); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	// set default values for configs
	config := Config{
		Extensions: make(map[string]ExtensionConfig),
	}

	if err := json.Unmarshal(configBytes, &config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	config.path = configPath

	return config, nil
}

func (c Config) Save() error {
	f, err := os.Create(c.path)
	if err != nil {
		return fmt.Errorf("failed to open config: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
