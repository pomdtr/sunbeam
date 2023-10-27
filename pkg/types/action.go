package types

type CommandRef struct {
	Extension string         `json:"extension,omitempty"`
	Command   string         `json:"command,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
}

type Action struct {
	Title string      `json:"title,omitempty"`
	Key   string      `json:"key,omitempty"`
	Type  CommandType `json:"type,omitempty"`

	Text string `json:"text,omitempty"`

	App    *Application `json:"app,omitempty"`
	Target string       `json:"target,omitempty"`

	Exit bool `json:"exit,omitempty"`

	Reload bool `json:"reload,omitempty"`

	Extension string         `json:"extension,omitempty"`
	Command   string         `json:"command,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
}

type CommandType string

const (
	CommandTypeRun    CommandType = "run"
	CommandTypeOpen   CommandType = "open"
	CommandTypeCopy   CommandType = "copy"
	CommandTypeReload CommandType = "reload"
	CommandTypeExit   CommandType = "exit"
)

type Application struct {
	Windows string `json:"windows,omitempty"`
	Mac     string `json:"mac,omitempty"`
	Linux   string `json:"linux,omitempty"`
}

type CommandInput struct {
	Params map[string]any `json:"params"`
	Query  string         `json:"query,omitempty"`
	Cwd    string         `json:"cwd,omitempty"`
}
