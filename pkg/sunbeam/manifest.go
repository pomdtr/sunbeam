package sunbeam

type Manifest struct {
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Preferences []Input       `json:"preferences,omitempty"`
	Commands    []CommandSpec `json:"commands"`
}

type CommandSpec struct {
	Name   string      `json:"name"`
	Title  string      `json:"title"`
	Hidden bool        `json:"hidden,omitempty"`
	Params []Input     `json:"params,omitempty"`
	Mode   CommandMode `json:"mode,omitempty"`
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

type CommandMode string

const (
	CommandModeSearch CommandMode = "search"
	CommandModeFilter CommandMode = "filter"
	CommandModeDetail CommandMode = "detail"
	CommandModeTTY    CommandMode = "tty"
	CommandModeSilent CommandMode = "silent"
)

type InputType string

const (
	InputString  InputType = "string"
	InputBoolean InputType = "boolean"
	InputNumber  InputType = "number"
)

type Input struct {
	Type     InputType `json:"type"`
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	Optional bool      `json:"optional,omitempty"`
	Default  any       `json:"default,omitempty"`
}
