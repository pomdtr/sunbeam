package pkg

import (
	"encoding/json"
	"fmt"
)

type Manifest struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Origin      string     `json:"origin,omitempty"`
	Entrypoint  Entrypoint `json:"entrypoint,omitempty"`
	Commands    []Command  `json:"commands"`
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
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Mode        CommandMode     `json:"mode"`
	Description string          `json:"description,omitempty"`
	Params      []CommandParams `json:"params,omitempty"`
}

type CommandMode string

const (
	CommandModeFilter    CommandMode = "filter"
	CommandModeGenerator CommandMode = "generator"
	CommandModeDetail    CommandMode = "detail"
	CommandModeText      CommandMode = "text"
	CommandModeSilent    CommandMode = "silent"
)

type CommandParams struct {
	Name        string    `json:"name"`
	Type        ParamType `json:"type"`
	Default     any       `json:"default,omitempty"`
	Optional    bool      `json:"optional,omitempty"`
	Description string    `json:"description,omitempty"`
}

type ParamType string

const (
	ParamTypeString  ParamType = "string"
	ParamTypeBoolean ParamType = "boolean"
)

type CommandInput struct {
	Query  string         `json:"query,omitempty"`
	Params map[string]any `json:"params,omitempty"`
}
