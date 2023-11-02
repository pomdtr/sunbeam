package types

type Text struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	Placeholder string `json:"placeholder,omitempty"`
	Default     string `json:"default,omitempty"`
}

type TextArea struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	Placeholder string `json:"placeholder,omitempty"`
	Default     string `json:"default,omitempty"`
}

type Checkbox struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	Label       string `json:"label,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Default     bool   `json:"default,omitempty"`
}

type Select struct {
	Title       string         `json:"title"`
	Type        string         `json:"type"`
	Placeholder string         `json:"placeholder,omitempty"`
	Default     any            `json:"default,omitempty"`
	Items       []DropDownItem `json:"items,omitempty"`
}

type DropDownItem struct {
	Title string `json:"title"`
	Value any    `json:"value"`
}
