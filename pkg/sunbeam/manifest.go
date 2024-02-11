package sunbeam

type Manifest struct {
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Imports     map[string]string `json:"imports,omitempty"`
	Commands    []Command         `json:"commands"`
	Root        []Action          `json:"root,omitempty"`
}

type Command struct {
	Name   string      `json:"name"`
	Hidden bool        `json:"hidden,omitempty"`
	Title  string      `json:"title"`
	Mode   CommandMode `json:"mode,omitempty"`
	Params []Param     `json:"params,omitempty"`
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
	ParamString  InputType = "string"
	ParamBoolean InputType = "boolean"
	ParamNumber  InputType = "number"
)

type Param struct {
	Type     InputType `json:"type"`
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	Optional bool      `json:"optional,omitempty"`
}
