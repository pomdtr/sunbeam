package types

type Manifest struct {
	Title       string        `json:"title"`
	Platforms   []Platfom     `json:"platforms,omitempty"`
	Description string        `json:"description,omitempty"`
	Require     []Requirement `json:"requirements,omitempty"`
	Root        []RootItem    `json:"root,omitempty"`
	Commands    []CommandSpec `json:"commands"`
}

type Platfom string

const (
	PlatformLinux Platfom = "linux"
	PlatformMac   Platfom = "macos"
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
	Inputs []Input     `json:"params,omitempty"`
	Mode   CommandMode `json:"mode,omitempty"`
}

type CommandMode string

const (
	CommandModeList   CommandMode = "list"
	CommandModeDetail CommandMode = "detail"
	CommandModeTTY    CommandMode = "tty"
	CommandModeSilent CommandMode = "silent"
)

type Input struct {
	Type        InputType `json:"type"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Required    bool      `json:"required,omitempty"`
	Default     any       `json:"default,omitempty"`
	Placeholder string    `json:"placeholder,omitempty"`
	Label       string    `json:"label,omitempty"`
}

type InputType string

const (
	InputText     InputType = "text"
	InputTextArea InputType = "textarea"
	InputPassword InputType = "password"
	InputCheckbox InputType = "checkbox"
	InputNumber   InputType = "number"
)
