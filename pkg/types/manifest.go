package types

type Manifest struct {
	Title       string        `json:"title"`
	Platforms   []Platfom     `json:"platforms,omitempty"`
	Description string        `json:"description,omitempty"`
	Require     []Requirement `json:"requirements,omitempty"`
	Root        []RootItem    `json:"root,omitempty"`
	Commands    []CommandSpec `json:"commands"`
	Preferences []Param       `json:"preferences,omitempty"`
}

type Platfom string

const (
	PlatformWindows Platfom = "windows"
	PlatformLinux   Platfom = "linux"
	PlatformMac     Platfom = "macos"
)

type Requirement struct {
	Name string `json:"name"`
	Link string `json:"link,omitempty"`
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
	Default     any       `json:"default,omitempty"`
}

type ParamType string

const (
	ParamTypeString  ParamType = "string"
	ParamTypeBoolean ParamType = "boolean"
	ParamTypeNumber  ParamType = "number"
)
