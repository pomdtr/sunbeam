package tui

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const keyringUser = "Sunbeam"
const keyringService = "Sunbeam Preferences"

type KeyStore struct {
	preferenceMap map[string]ScriptPreference
}

func LoadKeyStore() (*KeyStore, error) {
	var err error
	keyringValue, err := keyring.Get(keyringService, keyringUser)
	if errors.Is(err, keyring.ErrNotFound) {
		keyringValue = "{}"
	} else if err != nil {
		return nil, fmt.Errorf("failed to load keyring: %w", err)
	}

	preferenceMap := make(map[string]ScriptPreference)
	err = json.Unmarshal([]byte(keyringValue), &preferenceMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse keyring: %w", err)
	}

	return &KeyStore{
		preferenceMap: preferenceMap,
	}, nil
}

func GetPreferenceId(extension string, script string, name string) string {
	if script != "" {
		return fmt.Sprintf("%s.%s.%s", extension, script, name)
	}
	return fmt.Sprintf("%s.%s", extension, name)
}

type ScriptPreference struct {
	Name      string `json:"name"`
	Script    string `json:"script"`
	Extension string `json:"extension"`
	Value     any    `json:"value"`
}

func (k KeyStore) GetPreference(extension string, script string, name string) (ScriptPreference, bool) {
	if k.preferenceMap == nil {
		return ScriptPreference{}, false
	}
	scriptId := GetPreferenceId(extension, script, name)
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
		k.preferenceMap[GetPreferenceId(preference.Extension, preference.Script, preference.Name)] = preference
	}

	preferencesJSON, err := json.Marshal(k.preferenceMap)
	if err != nil {
		return err
	}

	return keyring.Set(keyringService, keyringUser, string(preferencesJSON))
}
