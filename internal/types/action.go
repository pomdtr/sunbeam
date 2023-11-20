package types

import (
	"encoding/json"
	"fmt"
)

type Action struct {
	Title string     `json:"title,omitempty"`
	Key   string     `json:"key,omitempty"`
	Type  ActionType `json:"type,omitempty"`

	Text string `json:"text,omitempty"`

	App    *Application `json:"app,omitempty"`
	Target string       `json:"target,omitempty"`

	Exit bool `json:"exit,omitempty"`

	Reload bool `json:"reload,omitempty"`

	Extension string           `json:"extension,omitempty"`
	Command   string           `json:"command,omitempty"`
	Params    map[string]Param `json:"params,omitempty"`

	Dir string `json:"dir,omitempty"`
}

type Param struct {
	Value    any  `json:"value,omitempty"`
	Default  any  `json:"default,omitempty"`
	Required bool `json:"required,omitempty"`
}

func (p *Param) UnmarshalJSON(bts []byte) error {
	var s string
	if err := json.Unmarshal(bts, &s); err == nil {
		p.Value = s
		return nil
	}

	var b bool
	if err := json.Unmarshal(bts, &b); err == nil {
		p.Value = b
		return nil
	}

	var n int
	if err := json.Unmarshal(bts, &n); err == nil {
		p.Value = n
		return nil
	}

	var param struct {
		Default  any  `json:"default,omitempty"`
		Required bool `json:"required,omitempty"`
	}

	if err := json.Unmarshal(bts, &param); err == nil {
		p.Default = param.Default
		p.Required = param.Required
		return nil
	}

	return fmt.Errorf("invalid param: %s", string(bts))
}

func (p Param) MarshalJSON() ([]byte, error) {
	if p.Value != nil {
		return json.Marshal(p.Value)
	}

	return json.Marshal(struct {
		Default  any  `json:"default,omitempty"`
		Required bool `json:"required,omitempty"`
	}{
		Default:  p.Default,
		Required: p.Required,
	})
}

type ActionType string

const (
	ActionTypeRun    ActionType = "run"
	ActionTypeOpen   ActionType = "open"
	ActionTypeCopy   ActionType = "copy"
	ActionTypeReload ActionType = "reload"
	ActionTypeEdit   ActionType = "edit"
	ActionTypeExec   ActionType = "exec"
	ActionTypeExit   ActionType = "exit"
	ActionTypeConfig ActionType = "config"
)

type Application struct {
	Macos string `json:"macos,omitempty"`
	Linux string `json:"linux,omitempty"`
}

type Payload struct {
	Command     string         `json:"command"`
	Preferences map[string]any `json:"preferences"`
	Params      map[string]any `json:"params"`
	Cwd         string         `json:"cwd"`
	Query       string         `json:"query,omitempty"`
}
