package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/pkg/schemas"
)

type Config struct {
	Oneliners  map[string]string            `json:"oneliners,omitempty"`
	Extensions map[string]extensions.Config `json:"extensions,omitempty"`
}

func (cfg Config) Aliases() []string {
	var aliases []string
	for alias := range cfg.Extensions {
		aliases = append(aliases, alias)
	}

	return aliases
}

func Dir() string {
	if env, ok := os.LookupEnv("SUNBEAM_CONFIG_DIR"); ok {
		return env
	}

	if env, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		return filepath.Join(env, "sunbeam")
	}

	return filepath.Join(os.Getenv("HOME"), ".config", "sunbeam")
}

func Path() string {
	configDir := Dir()
	if _, err := os.Stat(filepath.Join(configDir, "sunbeamrc")); err == nil {
		return filepath.Join(configDir, "sunbeamrc")
	}

	return filepath.Join(configDir, "sunbeam.json")
}

//go:embed sunbeam.json
var defaultConfigBytes []byte

func LoadBytes(configDir string) ([]byte, error) {
	if info, err := os.Stat(filepath.Join(configDir, "sunbeamrc")); err == nil {
		configPath := filepath.Join(configDir, info.Name())
		if err := os.Chmod(configPath, 0700); err != nil {
			return nil, fmt.Errorf("failed to chmod config: %w", err)
		}

		return exec.Command(configPath).Output()
	}

	if info, err := os.Stat(filepath.Join(configDir, "sunbeam.json")); err == nil {
		configPath := filepath.Join(configDir, info.Name())
		return os.ReadFile(configPath)
	}

	return nil, fmt.Errorf("failed to load config: %w", os.ErrNotExist)
}

func Load() (Config, error) {
	configDir := Dir()
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return Config{}, fmt.Errorf("failed to create config dir: %w", err)
	}

	configBytes, err := LoadBytes(configDir)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(filepath.Join(configDir, "sunbeam.json"), defaultConfigBytes, 0600); err != nil {
			return Config{}, fmt.Errorf("failed to write default config: %w", err)
		}

		configBytes = defaultConfigBytes
	} else if err != nil {
		return Config{}, fmt.Errorf("failed to load config: %w", err)
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
