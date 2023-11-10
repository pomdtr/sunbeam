package utils

import (
	"encoding/json"

	"github.com/99designs/keyring"
)

const (
	keyringServiceName = "sunbeam"
	KeyringKey         = "sunbeam"
	keyringLabel       = "Sunbeam"
)

type Keyring struct {
	keyring keyring.Keyring
	values  map[string]map[string]any
}

func LoadKeyring() (Keyring, error) {
	kv, err := keyring.Open(keyring.Config{
		ServiceName: keyringServiceName,
	})
	if err != nil {
		return Keyring{}, err
	}

	prefBytes, err := kv.Get(KeyringKey)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return Keyring{
				keyring: kv,
				values:  make(map[string]map[string]any),
			}, nil
		}
	}

	var sunbeamPrefs map[string]map[string]any
	if err := json.Unmarshal(prefBytes.Data, &sunbeamPrefs); err != nil {
		return Keyring{}, err
	}

	return Keyring{
		keyring: kv,
		values:  sunbeamPrefs,
	}, nil
}

func (p Keyring) Get(alias string) (map[string]any, bool) {
	values, ok := p.values[alias]
	return values, ok
}

func (p Keyring) Save(alias string, values map[string]any) error {
	p.values[alias] = values

	prefBytes, err := json.Marshal(p.values)
	if err != nil {
		return err
	}

	return p.keyring.Set(keyring.Item{
		Key:         KeyringKey,
		Data:        prefBytes,
		Label:       keyringLabel,
		Description: "Sunbeam preferences",
	})
}
