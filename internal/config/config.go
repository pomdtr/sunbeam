package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/types"
)

type Config struct {
	Root []types.RootItem  `json:"root"`
	Env  map[string]string `json:"env"`
}

var Path = filepath.Join(utils.ConfigHome(), "config.json")

func Load() (Config, error) {
	configPath := Path
	if _, err := os.Stat(configPath); err != nil {
		return Config{}, nil
	}

	var configBytes []byte
	bts, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}
	configBytes = bts

	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}
