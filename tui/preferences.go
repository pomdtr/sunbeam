package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type KeyStore struct {
	preferencePath string
	preferenceMap  map[string]ScriptPreference
}

func LoadKeyStore(preferencePath string) (*KeyStore, error) {
	preferenceMap := make(map[string]ScriptPreference)
	if _, err := os.Stat(preferencePath); os.IsNotExist(err) {
		return &KeyStore{
			preferencePath: preferencePath,
			preferenceMap:  preferenceMap,
		}, nil
	}

	preferenceBytes, err := os.ReadFile(preferencePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read preferences: %w", err)
	}
	err = json.Unmarshal(preferenceBytes, &preferenceMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse preferences: %w", err)
	}

	return &KeyStore{
		preferencePath: preferencePath,
		preferenceMap:  preferenceMap,
	}, nil
}

func (k *KeyStore) Save() (err error) {
	if _, err := os.Stat(path.Dir(k.preferencePath)); os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(k.preferencePath), 0755)
		if err != nil {
			return fmt.Errorf("failed to create preferences directory: %w", err)
		}
	}

	preferencesJSON, err := json.Marshal(k.preferenceMap)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	err = os.WriteFile(k.preferencePath, preferencesJSON, 0644)
	if err != nil {
		return fmt.Errorf("failed to write preferences: %w", err)
	}

	return nil
}

type ScriptPreference struct {
	Name    string
	Command string
	Value   any
}

func (k KeyStore) GetPreference(extension string, command string, name string) (any, bool) {
	for _, preference := range k.preferenceMap {
		if preference.Command == command && preference.Name == name {
			return preference.Value, true
		}
	}

	for _, preference := range k.preferenceMap {
		if preference.Command == "" && preference.Name == name {
			return preference.Value, true
		}
	}

	return nil, false
}

func (k *KeyStore) SetPreference(extensionName string, pref ScriptPreference) {
	k.preferenceMap[extensionName] = pref
}
