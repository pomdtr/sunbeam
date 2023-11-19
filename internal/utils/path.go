package utils

import (
	"os"
	"path/filepath"
)

func ConfigDir() string {
	if env, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		return filepath.Join(env, "sunbeam")
	}

	return filepath.Join(os.Getenv("HOME"), ".config", "sunbeam")
}
