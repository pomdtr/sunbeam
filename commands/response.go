package commands

type FormResponse struct {
	Title  string     `json:"title"`
	Method string     `json:"dest" validate:"oneof=args stdin"`
	Items  []FormItem `json:"items"`
}

type FormItem struct {
	Type    string `json:"type" validate:"required,oneof=text password"`
	Id      string `json:"id" validate:"required"`
	Name    string `json:"name" validate:"required"`
	Default string `json:"value"`
}

type ListItem struct {
	Icon     string         `json:"icon"`
	Title    string         `json:"title" validate:"required"`
	Subtitle string         `json:"subtitle"`
	Fill     string         `json:"fill"`
	Actions  []ScriptAction `json:"actions" validate:"required,gte=1,dive"`
}

type ScriptAction struct {
	Type    string         `json:"type" validate:"required,oneof=copy open url push"`
	Title   string         `json:"title"`
	Path    string         `json:"path"`
	Keybind string         `json:"keybind"`
	Params  map[string]any `json:"params"`
	Target  string         `json:"target,omitempty"`
	Command []string       `json:"command,omitempty"`
	Url     string         `json:"url,omitempty"`
	Content string         `json:"content,omitempty"`
}
