package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
)

//go:embed config.json
var configBytes []byte

type Config struct {
	Schema     string                     `json:"$schema,omitempty"`
	Oneliners  []Oneliner                 `json:"oneliners,omitempty"`
	Extensions map[string]ExtensionConfig `json:"extensions,omitempty"`
}

type ExtensionConfig struct {
	Origin      string            `json:"origin,omitempty"`
	Preferences types.Preferences `json:"preferences,omitempty"`
	Root        []types.RootItem  `json:"root,omitempty"`
}

func (e *ExtensionConfig) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		e.Origin = s
		return nil
	}

	var extensionRef struct {
		Origin      string         `json:"origin,omitempty"`
		Preferences map[string]any `json:"preferences,omitempty"`
		Root        []types.RootItem
	}

	if err := json.Unmarshal(b, &extensionRef); err == nil {
		e.Origin = extensionRef.Origin
		e.Preferences = extensionRef.Preferences
		e.Root = extensionRef.Root
		return nil
	}

	return fmt.Errorf("invalid extension ref: %s", string(b))
}

func (cfg Config) Aliases() []string {
	var aliases []string
	for alias := range cfg.Extensions {
		aliases = append(aliases, alias)
	}

	return aliases
}

type Oneliner struct {
	Title   string `json:"title"`
	Command string `json:"command"`
	Cwd     string `json:"cwd"`
	Exit    bool   `json:"exit"`
}

func Path() string {
	if env, ok := os.LookupEnv("SUNBEAM_CONFIG"); ok {
		return env
	}

	return filepath.Join(utils.ConfigHome(), "config.json")

}

func Load() (Config, error) {
	configPath := Path()
	if _, err := os.Stat(configPath); err != nil {

		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return Config{}, err
		}

		f, err := os.Create(configPath)
		if err != nil {
			return Config{}, err
		}
		defer f.Close()

		if _, err := f.Write(configBytes); err != nil {
			return Config{}, err
		}
	}

	var configBytes []byte
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	if err := schemas.ValidateConfig(configBytes); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}
