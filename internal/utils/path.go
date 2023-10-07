package utils

import (
	"os"
	"path/filepath"
)

func ConfigPath() string {
	if env, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		if _, err := os.Stat(filepath.Join(env, "sunbeam", "config.json")); err == nil {
			return filepath.Join(env, "sunbeam", "config.jsonc")
		} else {
			return filepath.Join(env, "sunbeam", "config.json")
		}
	}

	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "config.jsonc")); err == nil {
		return filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "config.jsonc")
	} else {
		return filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "config.json")
	}
}
