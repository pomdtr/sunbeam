package tui

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/pomdtr/sunbeam/app"
	"github.com/zalando/go-keyring"
)

const keyringUser = "Sunbeam"
const keyringService = "Sunbeam Preferences"

func getPreferenceMap() (map[string]app.ScriptPreference, error) {
	var err error
	keyringValue, err := keyring.Get(keyringService, keyringUser)
	log.Println("keyringValue", keyringValue)
	if errors.Is(err, keyring.ErrNotFound) {
		keyringValue = "{}"
	} else if err != nil {
		return nil, err
	}

	preferenceMap := make(map[string]app.ScriptPreference)
	err = json.Unmarshal([]byte(keyringValue), &preferenceMap)
	if err != nil {
		return nil, err
	}

	return preferenceMap, nil
}

func getPreferences() ([]app.ScriptPreference, error) {
	preferenceMap, err := getPreferenceMap()
	if err != nil {
		return nil, err
	}

	preferences := make([]app.ScriptPreference, 0, len(preferenceMap))
	for _, preference := range preferenceMap {
		preferences = append(preferences, preference)
	}

	return preferences, nil
}

func setPreference(preferences ...app.ScriptPreference) error {
	preferenceMap, err := getPreferenceMap()
	if err != nil {
		return err
	}

	for _, preference := range preferences {
		preferenceMap[preference.Id()] = preference
	}

	preferencesJSON, err := json.Marshal(preferenceMap)
	if err != nil {
		return err
	}

	return keyring.Set(keyringService, keyringUser, string(preferencesJSON))
}
