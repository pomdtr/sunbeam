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
	store map[string]ScriptPreference
}

var keystore KeyStore

func init() {
	keystore = KeyStore{}
	keystore.Init()
}

func GetPreferenceId(extension string, script string, name string) string {
	if script != "" {
		return fmt.Sprintf("%s.%s.%s", extension, script, name)
	}
	return fmt.Sprintf("%s.%s", extension, name)
}

func (k *KeyStore) Init() error {
	var err error
	keyringValue, err := keyring.Get(keyringService, keyringUser)
	if errors.Is(err, keyring.ErrNotFound) {
		keyringValue = "{}"
	} else if err != nil {
		return err
	}

	preferenceMap := make(map[string]ScriptPreference)
	err = json.Unmarshal([]byte(keyringValue), &preferenceMap)
	if err != nil {
		return err
	}

	return nil
}

type ScriptPreference struct {
	Name      string `json:"name"`
	Script    string `json:"script"`
	Extension string `json:"extension"`
	Value     any    `json:"value"`
}

func (k KeyStore) GetPreference(extension string, script string, name string) (ScriptPreference, bool) {
	if k.store == nil {
		return ScriptPreference{}, false
	}
	scriptId := GetPreferenceId(extension, script, name)
	if preference, ok := k.store[scriptId]; ok {
		return preference, true
	}

	extensionId := GetPreferenceId(extension, "", name)
	if preference, ok := k.store[extensionId]; ok {
		return preference, ok
	}

	return ScriptPreference{}, false
}

func (k *KeyStore) SetPreference(preferences ...ScriptPreference) error {
	for _, preference := range preferences {
		k.store[GetPreferenceId(preference.Extension, preference.Script, preference.Name)] = preference
	}

	preferencesJSON, err := json.Marshal(k.store)
	if err != nil {
		return err
	}

	return keyring.Set(keyringService, keyringUser, string(preferencesJSON))
}
