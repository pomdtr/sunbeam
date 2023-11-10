package utils

import (
	"encoding/json"
	"fmt"

	"github.com/99designs/keyring"
)

func LoadPreferences(alias string) (map[string]any, error) {
	kv, err := keyring.Open(keyring.Config{
		ServiceName: "sunbeam",
	})
	if err != nil {
		return nil, err
	}

	v, err := kv.Get("preferences")
	if err != nil {
		return nil, err
	}

	var sunbeamPrefs map[string]map[string]any
	if err := json.Unmarshal(v.Data, &sunbeamPrefs); err != nil {
		return nil, err
	}

	extensionPref, ok := sunbeamPrefs[alias]
	if !ok {
		return nil, fmt.Errorf("no preferences found for %s", alias)
	}

	return extensionPref, nil
}

func SavePrefs(alias string, prefs map[string]any) error {
	sunbeamPrefs, err := LoadPreferences(alias)
	if err != nil {
		sunbeamPrefs = make(map[string]any)
	}

	sunbeamPrefs[alias] = prefs

	kv, err := keyring.Open(keyring.Config{
		ServiceName: "sunbeam",
	})
	if err != nil {
		return err
	}

	prefsBytes, err := json.Marshal(sunbeamPrefs)
	if err != nil {
		return err
	}

	return kv.Set(keyring.Item{
		Key:   "preferences",
		Label: "Sunbeam Preferences",
		Data:  prefsBytes,
	})
}
