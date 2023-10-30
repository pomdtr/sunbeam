package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/tailscale/hujson"
)

type Config struct {
	Root []types.RootItem  `json:"root"`
	Env  map[string]string `json:"env"`
}

func Load() (Config, error) {
	var configBytes []byte
	if _, err := os.Stat(filepath.Join(utils.ConfigHome(), "config.json")); err == nil {
		bts, err := os.ReadFile(filepath.Join(utils.ConfigHome(), "config.json"))
		if err != nil {
			return Config{}, err
		}
		configBytes = bts
	} else if _, err := os.Stat(filepath.Join(utils.ConfigHome(), "config.jsonc")); err == nil {
		bts, err := os.ReadFile(filepath.Join(utils.ConfigHome(), "config.jsonc"))
		if err != nil {
			return Config{}, err
		}
		jsonBytes, err := hujson.Standardize(bts)
		if err != nil {
			return Config{}, err
		}
		configBytes = jsonBytes
	} else {
		return Config{
			Root: []types.RootItem{},
			Env:  map[string]string{},
		}, nil
	}

	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}
