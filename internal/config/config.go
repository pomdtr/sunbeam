package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pomdtr/sunbeam/internal/schemas"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/tailscale/hujson"
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
	Env        map[string]string          `json:"env,omitempty"`
	Extensions map[string]ExtensionConfig `json:"extensions,omitempty"`
	Root       []sunbeam.Action           `json:"root,omitempty"`
	Path       string                     `json:"-"`
}

type ExtensionConfig struct {
	Origin string            `json:"origin"`
	Env    map[string]string `json:"env,omitempty"`
}

func (cfg Config) Resolve(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(os.Getenv("HOME"), path[2:])
	}

	if !filepath.IsAbs(path) {
		return filepath.Join(filepath.Dir(cfg.Path), path)
	}

	return path
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

	configJsonBytes, err := hujson.Standardize(configBytes)
	if err != nil {
		return Config{}, fmt.Errorf("failed to standardize config: %w", err)
	}

	if err := schemas.ValidateConfig(configJsonBytes); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return Config{}, fmt.Errorf("failed to get executable path: %w", err)
	}

	// set default values for configs
	config := Config{
		Extensions: map[string]ExtensionConfig{
			"std": {
				Origin: execPath,
			},
		},
		Env: make(map[string]string),
	}

	if err := json.Unmarshal(configJsonBytes, &config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	config.Path = configPath

	return config, nil
}

func (c Config) Save() error {
	f, err := os.Create(c.Path)
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
