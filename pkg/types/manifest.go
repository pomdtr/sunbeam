package types

type Manifest struct {
	Title       string     `json:"title"`
	Homepage    string     `json:"homepage,omitempty"`
	Description string     `json:"description,omitempty"`
	Commands    []Command  `json:"commands"`
	Items       []RootItem `json:"items,omitempty"`
}

type RootItem struct {
	Title     string         `json:"title"`
	Extension string         `json:"extension"`
	Origin    string         `json:"-"`
	Command   string         `json:"command"`
	Params    map[string]any `json:"params,omitempty"`
}

type Command struct {
	Name        string         `json:"name"`
	Title       string         `json:"title"`
	Mutation    bool           `json:"mutation,omitempty"`
	Description string         `json:"description,omitempty"`
	Params      []CommandParam `json:"params,omitempty"`
	Mode        CommandMode    `json:"mode,omitempty"`
}

type CommandMode string

const (
	CommandModePage   CommandMode = "page"
	CommandModeAction CommandMode = "action"
	CommandModeSilent CommandMode = "silent"
)

type CommandParam struct {
	Name        string    `json:"name"`
	Type        ParamType `json:"type"`
	Optional    bool      `json:"optional,omitempty"`
	Description string    `json:"description,omitempty"`
}

type ParamType string

const (
	ParamTypeString  ParamType = "string"
	ParamTypeBoolean ParamType = "boolean"
)
