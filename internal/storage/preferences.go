package preferences

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal/utils"
)

var PreferencesPath = filepath.Join(utils.DataHome(), "preferences.json")

type Preferences map[string]Item

type Item struct {
	Origin      string         `json:"origin"`
	Preferences map[string]any `json:"preferences"`
}

func Load(alias string, origin string) (map[string]any, error) {
	bts, err := os.ReadFile(PreferencesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}

		return nil, err
	}

	var preferences Preferences
	if err := json.Unmarshal(bts, &preferences); err != nil {
		return nil, err
	}

	item, ok := preferences[alias]
	if !ok {
		return make(map[string]any), nil
	}

	if item.Origin != origin {
		return make(map[string]any), nil
	}

	return item.Preferences, nil
}

func Save(alias string, origin string, values map[string]any) error {
	inputBts, err := os.ReadFile(PreferencesPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		inputBts = []byte("{}")
	}

	var preferences Preferences
	if err := json.Unmarshal(inputBts, &preferences); err != nil {
		return err
	}

	preferences[alias] = Item{
		Origin:      origin,
		Preferences: values,
	}

	if err := os.MkdirAll(filepath.Dir(PreferencesPath), 0600); err != nil {
		return err
	}

	bts, err := json.MarshalIndent(preferences, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(PreferencesPath, bts, 0644)
}
