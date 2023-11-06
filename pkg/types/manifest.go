package types

type Manifest struct {
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Require     []Requirement `json:"requirements,omitempty"`
	Env         []Env         `json:"env,omitempty"`
	Root        []RootItem    `json:"root,omitempty"`
	Commands    []CommandSpec `json:"commands"`
}

type Requirement struct {
	Name string `json:"name"`
	Link string `json:"link,omitempty"`
}

type Env struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type RootItem struct {
	Title   string         `json:"title"`
	Command string         `json:"command"`
	Params  map[string]any `json:"params"`
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
	CommandModeList   CommandMode = "list"
	CommandModeDetail CommandMode = "detail"
	CommandModeTTY    CommandMode = "tty"
	CommandModeSilent CommandMode = "silent"
)

type Param struct {
	Name        string    `json:"name"`
	Type        ParamType `json:"type"`
	Description string    `json:"description,omitempty"`
	Required    bool      `json:"required,omitempty"`
}

type ParamType string

const (
	ParamTypeString  ParamType = "string"
	ParamTypeBoolean ParamType = "boolean"
	ParamTypeNumber  ParamType = "number"
)
