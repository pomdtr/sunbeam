package tui

import (
	"encoding/json"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sunbeamlauncher/sunbeam/app"
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

var keyStore *KeyStore

func init() {
	var err error
	keyStore, err = LoadKeyStore()
	if err != nil {
		panic(err)
	}
}

type PreferenceForm struct {
	extension    app.Extension
	onSuccessCmd tea.Cmd
	script       app.Script
	*Form
}

func NewPreferenceForm(extension app.Extension, script app.Script) *PreferenceForm {
	formitems := make([]FormItem, 0)
	for _, preference := range extension.Preferences {
		if prefValue, ok := keyStore.GetPreference(extension.Name, "", preference.Name); ok {
			preference.Default.Value = prefValue.Value
		}

		formitems = append(formitems, NewFormItem(preference))
	}

	for _, preference := range script.Preferences {
		if prefValue, ok := keyStore.GetPreference(extension.Name, script.Name, preference.Name); ok {
			preference.Default.Value = prefValue.Value
		}

		formitems = append(formitems, NewFormItem(preference))
	}

	return &PreferenceForm{
		extension:    extension,
		script:       script,
		Form:         NewForm(fmt.Sprintf("%s · Preferences", extension.Title), "Preferences", formitems),
		onSuccessCmd: PopCmd,
	}
}

func (p *PreferenceForm) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case SubmitMsg:
		preferences := make([]ScriptPreference, 0)
		for _, input := range p.extension.Preferences {
			value, ok := msg.Values[input.Name]
			if !ok {
				continue
			}
			preference := ScriptPreference{
				Name:      input.Name,
				Value:     value,
				Extension: p.extension.Name,
			}
			preferences = append(preferences, preference)
		}

		for _, input := range p.script.Preferences {
			value, ok := msg.Values[input.Name]
			if !ok {
				continue
			}
			preference := ScriptPreference{
				Name:      input.Name,
				Value:     value,
				Extension: p.extension.Name,
				Script:    p.script.Name,
			}
			preferences = append(preferences, preference)
		}

		err := keyStore.SetPreference(preferences...)
		if err != nil {
			return p, NewErrorCmd(err)
		}

		return p, p.onSuccessCmd
	}

	page, cmd := p.Form.Update(msg)
	p.Form = page.(*Form)

	return p, cmd
}
