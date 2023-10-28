package types

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

type CommandRef struct {
	Extension string         `json:"extension,omitempty"`
	Command   string         `json:"command,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
}

type Action struct {
	Title string     `json:"title,omitempty"`
	Key   string     `json:"key,omitempty"`
	Type  ActionType `json:"type,omitempty"`

	Text string `json:"text,omitempty"`

	App    *Application `json:"app,omitempty"`
	Target string       `json:"target,omitempty"`

	Exit bool `json:"exit,omitempty"`

	Reload bool `json:"reload,omitempty"`

	Extension string `json:"extension,omitempty"`
	Command   string `json:"command,omitempty"`
	Params    Params `json:"params,omitempty"`
}

type Params map[string]any

func (p *Params) UnmarshalJSON(b []byte) error {
	var params map[string]any
	if err := json.Unmarshal(b, &params); err != nil {
		return err
	}

	for k, v := range params {
		v, ok := v.(map[string]any)
		if !ok {
			continue
		}

		if _, ok := v["type"]; !ok {
			continue
		}

		vType, ok := v["type"].(string)
		if !ok {
			continue
		}

		switch vType {
		case "text":
			var text Text
			if err := mapstructure.Decode(v, &text); err != nil {
				return err
			}

			params[k] = text
		case "textarea":
			var textarea TextArea
			if err := mapstructure.Decode(v, &textarea); err != nil {
				return err
			}
			params[k] = textarea
		case "checkbox":
			var checkbox Checkbox
			if err := mapstructure.Decode(v, &checkbox); err != nil {
				return err
			}
			params[k] = checkbox
		case "select":
			var selectItem Select
			if err := mapstructure.Decode(v, &selectItem); err != nil {
				return err
			}
			params[k] = selectItem
		}
	}

	*p = params
	return nil
}

type ActionType string

const (
	ActionTypeRun    ActionType = "run"
	ActionTypeOpen   ActionType = "open"
	ActionTypeCopy   ActionType = "copy"
	ActionTypeReload ActionType = "reload"
	ActionTypeExit   ActionType = "exit"
)

type Application struct {
	Windows string `json:"windows,omitempty"`
	Mac     string `json:"mac,omitempty"`
	Linux   string `json:"linux,omitempty"`
}

type CommandInput struct {
	Command string         `json:"command"`
	Params  map[string]any `json:"params"`
	Query   string         `json:"query,omitempty"`
}
