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

func CacheDir() string {
	if env, ok := os.LookupEnv("XDG_CACHE_HOME"); ok {
		return filepath.Join(env, "sunbeam")
	}

	return filepath.Join(os.Getenv("HOME"), ".cache", "sunbeam")
}

func ExtensionsDir() string {
	if env, ok := os.LookupEnv("SUNBEAM_DIR"); ok {
		return env
	}

	if env, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		return filepath.Join(env, "sunbeam", "extensions")
	}

	return filepath.Join(os.Getenv("HOME"), ".local", "share", "sunbeam", "extensions")
}

func DataDir() string {
	if env, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		return filepath.Join(env, "sunbeam")
	}

	return filepath.Join(os.Getenv("HOME"), ".local", "share", "sunbeam")
}
