package api

type ListItem struct {
	Icon     string         `json:"icon"`
	Title    string         `json:"title"`
	Subtitle string         `json:"subtitle"`
	Fill     string         `json:"fill"`
	Actions  []ScriptAction `json:"actions"`
}

type ScriptAction struct {
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	Path        string            `json:"path"`
	Keybind     string            `json:"keybind"`
	Params      map[string]string `json:"params"`
	Target      string            `json:"target,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Application string            `json:"application,omitempty"`
	Url         string            `json:"url,omitempty"`
	Content     string            `json:"content,omitempty"`
}
