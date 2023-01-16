package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/app"
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
	if os.Stat(path.Dir(k.preferencePath)); os.IsNotExist(err) {
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

	return keyStore.Save()
}

var keyStore *KeyStore

// TODO: move this to the root model init function
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

type PreferenceForm struct {
	extension    app.Extension
	onSuccessCmd tea.Cmd
	script       app.Command
	*Form
}

func NewPreferenceForm(extension app.Extension, script app.Command) *PreferenceForm {
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
