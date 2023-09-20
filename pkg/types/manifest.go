package types

import (
	"encoding/json"
	"fmt"
)

type Manifest struct {
	Title       string             `json:"title"`
	Homepage    string             `json:"homepage,omitempty"`
	Description string             `json:"description,omitempty"`
	Commands    map[string]Command `json:"commands"`
	Items       []RootItem         `json:"items,omitempty"`
}

type RootItem struct {
	Title   string         `json:"title,omitempty"`
	Command string         `json:"command"`
	Params  map[string]any `json:"params,omitempty"`
}

type Entrypoint []string

func (e *Entrypoint) UnmarshalJSON(b []byte) error {
	var entrypoint string
	if err := json.Unmarshal(b, &entrypoint); err == nil {
		*e = Entrypoint{entrypoint}
		return nil
	}

	var entrypoints []string
	if err := json.Unmarshal(b, &entrypoints); err == nil {
		*e = Entrypoint(entrypoints)
		return nil
	}

	return fmt.Errorf("invalid entrypoint: %s", string(b))
}

type Command struct {
	Title       string                  `json:"title"`
	Mutation    bool                    `json:"mutation,omitempty"`
	Description string                  `json:"description,omitempty"`
	Params      map[string]CommandParam `json:"params,omitempty"`
	Mode        CommandMode             `json:"mode,omitempty"`
}

type CommandMode string

const (
	CommandModePage   CommandMode = "page"
	CommandModeAction CommandMode = "action"
	CommandModeSilent CommandMode = "silent"
)

type CommandParam struct {
	Type        ParamType `json:"type"`
	Optional    bool      `json:"optional,omitempty"`
	Description string    `json:"description,omitempty"`
}

type ParamType string

const (
	ParamTypeString  ParamType = "string"
	ParamTypeBoolean ParamType = "boolean"
)
