package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/tailscale/hujson"
)

type Config struct {
	Schema     string            `json:"$schema,omitempty"`
	Oneliners  []Oneliner        `json:"oneliners,omitempty"`
	Extensions map[string]string `json:"extensions,omitempty"`
}

func (cfg Config) Aliases() []string {
	var aliases []string
	for alias := range cfg.Extensions {
		aliases = append(aliases, alias)
	}

	return aliases
}

var DefaultConfig = Config{
	Schema: "https://github.com/pomdtr/sunbeam/releases/latest/download/config.schema.json",
	Oneliners: []Oneliner{
		{
			Title:   "Open Sunbeam Docs",
			Command: "sunbeam open https://pomdtr.github.io/sunbeam/introduction",
		},
		{
			Title:   "Edit Config",
			Command: "sunbeam edit --config",
		},
	},
	Extensions: map[string]string{
		"devdocs": "https://raw.githubusercontent.com/pomdtr/sunbeam/main/extensions/devdocs.sh",
		"google":  "https://raw.githubusercontent.com/pomdtr/sunbeam/main/extensions/google.sh",
	},
}

type Oneliner struct {
	Title   string `json:"title"`
	Command string `json:"command"`
}

func Path() string {
	if env, ok := os.LookupEnv("SUNBEAM_CONFIG"); ok {
		return env
	}

	if _, err := os.Stat(filepath.Join(utils.ConfigHome(), "config.jsonc")); err == nil {
		return filepath.Join(utils.ConfigHome(), "config.jsonc")
	}

	return filepath.Join(utils.ConfigHome(), "config.json")

}

func Load() (Config, error) {
	configPath := Path()
	if _, err := os.Stat(configPath); err != nil {
		configBytes, err := json.MarshalIndent(DefaultConfig, "", "  ")
		if err != nil {
			return Config{}, err
		}

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

		return DefaultConfig, nil
	}

	var configBytes []byte
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	if filepath.Ext(configPath) == ".jsonc" {
		bts, err := hujson.Standardize(configBytes)
		if err != nil {
			return Config{}, err
		}

		configBytes = bts
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
