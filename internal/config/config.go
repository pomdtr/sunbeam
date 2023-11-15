package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/tailscale/hujson"
)

//go:embed sunbeam.json
var configBytes []byte

type Config struct {
	Schema     string                       `json:"$schema,omitempty"`
	Oneliners  map[string]Oneliner          `json:"oneliners,omitempty"`
	Extensions map[string]extensions.Config `json:"extensions,omitempty"`
}

func (cfg Config) Aliases() []string {
	var aliases []string
	for alias := range cfg.Extensions {
		aliases = append(aliases, alias)
	}

	return aliases
}

type Oneliner struct {
	Command string `json:"command"`
	Cwd     string `json:"cwd"`
	Exit    bool   `json:"exit"`
}

func (o *Oneliner) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		o.Command = s
		return nil
	}

	type Alias Oneliner
	var alias Alias
	if err := json.Unmarshal(b, &alias); err == nil {
		o.Command = alias.Command
		o.Cwd = alias.Cwd
		o.Exit = alias.Exit
		return nil
	}

	return fmt.Errorf("invalid oneliner: %s", string(b))
}

func Path() string {
	if env, ok := os.LookupEnv("SUNBEAM_CONFIG"); ok {
		return env
	}

	if _, err := os.Stat(filepath.Join(utils.ConfigHome(), "config.jsonc")); err == nil {
		return filepath.Join(utils.ConfigHome(), "sunbeam.jsonc")
	}

	return filepath.Join(utils.ConfigHome(), "sunbeam.json")
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

	jsonBytes, err := hujson.Standardize(configBytes)
	if err != nil {
		return Config{}, err
	}

	if err := schemas.ValidateConfig(jsonBytes); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}
