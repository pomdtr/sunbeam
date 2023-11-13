package types

type CommandRef struct {
	Extension string         `json:"extension,omitempty"`
	Command   string         `json:"command,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
}

type Action struct {
	Title string     `json:"title,omitempty"`
	Key   string     `json:"key,omitempty"`
	Type  ActionType `json:"type,omitempty"`

	Text string `json:"text,omitempty"`

	App    *Application `json:"app,omitempty"`
	Target string       `json:"target,omitempty"`

	Exit bool `json:"exit,omitempty"`

	Reload bool `json:"reload,omitempty"`

	Extension string `json:"extension,omitempty"`
	Command   string `json:"command,omitempty"`
	Params    Params `json:"params,omitempty"`
}

type Params map[string]any

type ActionType string

const (
	ActionTypeRun    ActionType = "run"
	ActionTypeOpen   ActionType = "open"
	ActionTypeCopy   ActionType = "copy"
	ActionTypeReload ActionType = "reload"
	ActionTypeEdit   ActionType = "edit"
	ActionTypeExec   ActionType = "exec"
	ActionTypeExit   ActionType = "exit"
)

type Application struct {
	Windows string `json:"windows,omitempty"`
	Macos   string `json:"macos,omitempty"`
	Linux   string `json:"linux,omitempty"`
}

type Payload struct {
	Command     string         `json:"command"`
	Preferences Preferences    `json:"preferences"`
	Params      map[string]any `json:"params"`
	Cwd         string         `json:"cwd"`
	Query       string         `json:"query,omitempty"`
}

type Preferences map[string]any
