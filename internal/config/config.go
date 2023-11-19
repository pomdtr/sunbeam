package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/schemas"
	"github.com/pomdtr/sunbeam/internal/utils"
)

var Path string

func init() {
	if env, ok := os.LookupEnv("SUNBEAM_CONFIG_FILE"); ok {
		Path = env
	} else {
		Path = filepath.Join(utils.ConfigDir(), "sunbeam.json")
	}
}

type Config struct {
	Oneliners  []Oneliner                   `json:"oneliners,omitempty"`
	Extensions map[string]extensions.Config `json:"extensions,omitempty"`
	path       string                       `json:"-"`
}

type Oneliner struct {
	Title   string `json:"title,omitempty"`
	Command string `json:"command"`
	Dir     string `json:"dir,omitempty"`
	Exit    bool   `json:"exit,omitempty"`
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

	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	config.path = configPath

	return config, nil
}

func (c Config) Save() error {
	bts, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.path, bts, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
