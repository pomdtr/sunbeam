package types

type Manifest struct {
	Title       string        `json:"title"`
	Origin      string        `json:"origin,omitempty"`
	Version     string        `json:"version,omitempty"`
	Description string        `json:"description,omitempty"`
	Commands    []CommandSpec `json:"commands"`
}

type CommandSpec struct {
	Name   string      `json:"name"`
	Title  string      `json:"title"`
	Hidden bool        `json:"hidden,omitempty"`
	Params []Param     `json:"params,omitempty"`
	Mode   CommandMode `json:"mode,omitempty"`
}

type CommandMode string

const (
	CommandModeView   CommandMode = "view"
	CommandModeNoView CommandMode = "no-view"
	CommandModeTTY    CommandMode = "tty"
)

type Param struct {
	Name        string    `json:"name"`
	Type        ParamType `json:"type"`
	Optional    bool      `json:"optional,omitempty"`
	Description string    `json:"description,omitempty"`
}

type ParamType string

const (
	ParamTypeString  ParamType = "string"
	ParamTypeBoolean ParamType = "boolean"
	ParamTypeNumber  ParamType = "number"
)
