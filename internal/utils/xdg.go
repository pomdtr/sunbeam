package utils

import (
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/pkg/types"
)

type ExtensionCache map[string]types.Manifest

func DataHome() string {
	if env, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		return filepath.Join(env, "sunbeam")
	}

	homedir, _ := os.UserHomeDir()
	return filepath.Join(homedir, ".local", "share", "sunbeam")
}

func ConfigHome() string {
	if env, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		return filepath.Join(env, "sunbeam")
	}

	homedir, _ := os.UserHomeDir()
	return filepath.Join(homedir, ".config", "sunbeam")
}

func CacheHome() string {
	if env, ok := os.LookupEnv("XDG_CACHE_HOME"); ok {
		return filepath.Join(env, "sunbeam")
	}

	homedir, _ := os.UserHomeDir()
	return filepath.Join(homedir, ".cache", "sunbeam")
}
