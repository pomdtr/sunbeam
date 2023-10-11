package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/pkg/types"
)

type Config struct {
	Root map[string]types.Command `json:"root"`
}

type ExtensionCache map[string]types.Manifest

func dataHome() string {
	if env, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		return filepath.Join(env, "sunbeam")
	}

	return filepath.Join(os.Getenv("HOME"), ".local", "share", "sunbeam")
}

func cacheHome() string {
	if env, ok := os.LookupEnv("XDG_CACHE_HOME"); ok {
		return filepath.Join(env, "sunbeam")
	}

	return filepath.Join(os.Getenv("HOME"), ".cache", "sunbeam")
}

type History struct {
	entries map[string]int64
	path    string
}

func LoadHistory(fp string) (History, error) {
	f, err := os.Open(fp)
	if err != nil {
		return History{}, err
	}

	var entries map[string]int64
	if err := json.NewDecoder(f).Decode(&entries); err != nil {
		return History{}, err
	}

	return History{
		entries: entries,
		path:    fp,
	}, nil
}

func (h History) Save() error {
	f, err := os.OpenFile(h.path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(h.path), 0755); err != nil {
			return err
		}

		f, err = os.Create(h.path)
		if err != nil {
			return err
		}
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(h.entries); err != nil {
		return err
	}

	return nil
}
