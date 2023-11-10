package utils

import (
	"encoding/json"

	"github.com/99designs/keyring"
)

func LoadPreferences(alias string) (map[string]any, error) {
	kv, err := keyring.Open(keyring.Config{
		ServiceName: "sunbeam",
	})
	if err != nil {
		return nil, err
	}

	v, err := kv.Get(alias)
	if err != nil {
		return nil, err
	}

	var prefs map[string]any
	if err := json.Unmarshal(v.Data, &prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

func SavePrefs(alias string, prefs map[string]any) error {
	kv, err := keyring.Open(keyring.Config{
		ServiceName: "sunbeam",
	})
	if err != nil {
		return err
	}

	prefsBytes, err := json.Marshal(prefs)
	if err != nil {
		return err
	}

	return kv.Set(keyring.Item{
		Key:  alias,
		Data: prefsBytes,
	})
}
