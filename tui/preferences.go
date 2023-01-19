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

func GetPreferenceId(extension string, command string, name string) string {
	if command != "" {
		return fmt.Sprintf("%s.%s.%s", extension, command, name)
	}
	return fmt.Sprintf("%s.%s", extension, name)
}

type ScriptPreference struct {
	Name      string
	Command   string
	Extension string
	Value     any
}

func (k KeyStore) GetPreference(extension string, command string, name string) (ScriptPreference, bool) {
	if k.preferenceMap == nil {
		return ScriptPreference{}, false
	}
	scriptId := GetPreferenceId(extension, command, name)
	if preference, ok := k.preferenceMap[scriptId]; ok {
		return preference, true
	}

	extensionId := GetPreferenceId(extension, "", name)
	if preference, ok := k.preferenceMap[extensionId]; ok {
		return preference, ok
	}

	return ScriptPreference{}, false
}

func (k *KeyStore) SetPreference(preferences ...ScriptPreference) error {
	for _, preference := range preferences {
		k.preferenceMap[GetPreferenceId(preference.Extension, preference.Command, preference.Name)] = preference
	}

	return keyStore.Save()
}

// TODO: Remove this
var keyStore *KeyStore

func init() {
	var err error

	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	preferencePath := path.Join(homedir, ".config", "sunbeam", "preferences.json")
	keyStore, err = LoadKeyStore(preferencePath)
	if err != nil {
		panic(err)
	}
}
