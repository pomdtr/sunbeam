package api

type ListItem struct {
	Icon     string         `json:"icon"`
	Title    string         `json:"title" validate:"required"`
	Subtitle string         `json:"subtitle"`
	Fill     string         `json:"fill"`
	Actions  []ScriptAction `json:"actions" validate:"required,gte=1,dive"`
}

type ScriptAction struct {
	Type    string            `json:"type" validate:"required,oneof=copy open url push"`
	Title   string            `json:"title"`
	Path    string            `json:"path"`
	Keybind string            `json:"keybind"`
	Params  map[string]string `json:"params"`
	Target  string            `json:"target,omitempty"`
	Command []string          `json:"command,omitempty"`
	Url     string            `json:"url,omitempty"`
	Content string            `json:"content,omitempty"`
}
