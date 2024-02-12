package sunbeam

type Action struct {
	Title     string         `json:"title,omitempty"`
	Extension string         `json:"extension,omitempty"`
	Command   string         `json:"command,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
	Reload    bool           `json:"reload,omitempty"`
}

type Payload struct {
	Command string         `json:"command"`
	Params  map[string]any `json:"params"`
	Cwd     string         `json:"cwd"`
	Query   string         `json:"query,omitempty"`
}
