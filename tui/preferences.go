package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type KeyStore struct {
	preferencePath string
	preferenceMap  map[string]any
}

func LoadKeyStore(preferencePath string) (*KeyStore, error) {
	preferenceMap := make(map[string]any)
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
	preferenceId := k.PreferenceId(extension, command, name)
	if value, ok := k.preferenceMap[preferenceId]; ok {
		return value, true
	}

	preferenceId = k.PreferenceId(extension, "", name)
	if value, ok := k.preferenceMap[preferenceId]; ok {
		return value, true
	}

	return nil, false
}

func (k *KeyStore) PreferenceId(extension string, command string, name string) string {
	if command == "" {
		return fmt.Sprintf("%s.%s", extension, name)
	}
	return fmt.Sprintf("%s.%s.%s", extension, command, name)
}

func (k *KeyStore) SetPreference(extensionName string, commandName string, name string, value any) {
	preferenceId := k.PreferenceId(extensionName, commandName, name)
	k.preferenceMap[preferenceId] = value
}
