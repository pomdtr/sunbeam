package types

type Manifest struct {
	Title       string        `json:"title"`
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
	CommandModePage   CommandMode = "page"
	CommandModeSilent CommandMode = "silent"
)

type Param struct {
	Name        string    `json:"name"`
	Type        ParamType `json:"type"`
	Description string    `json:"description,omitempty"`
	Required    bool      `json:"required,omitempty"`
	Default     any       `json:"default,omitempty"`
}

type ParamType string

const (
	ParamTypeString  ParamType = "string"
	ParamTypeBoolean ParamType = "boolean"
	ParamTypeNumber  ParamType = "number"
)
