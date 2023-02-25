package app

type Page struct {
	Type  string `json:"type"`
	Title string `json:"title"`

	*Detail
	*List
}

type Detail struct {
	Content Preview  `json:"content,omitempty"`
	Actions []Action `json:"actions,omitempty"`
}

type List struct {
	ShowPreview   bool       `json:"showPreview,omitempty" yaml:"showPreview"`
	GenerateItems bool       `json:"generateItems,omitempty" yaml:"generateItems"`
	Items         []ListItem `json:"items"`
	EmptyView     struct {
		Text    string   `json:"text"`
		Actions []Action `json:"actions"`
	} `json:"emptyView,omitempty" yaml:"emptyView"`
}

type Preview struct {
	Text     string         `json:"text,omitempty"`
	Language string         `json:"language,omitempty"`
	Command  string         `json:"command,omitempty"`
	With     map[string]any `json:"with,omitempty"`
}

type ListItem struct {
	Id          string   `json:"id"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Preview     *Preview `json:"preview,omitempty"`
	Accessories []string `json:"accessories,omitempty"`
	Actions     []Action `json:"actions"`
}

type Action struct {
	Title string `json:"title"`
	Type  string `json:"type"`

	Text string `json:"text,omitempty"`

	Url  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`

	Command   string         `json:"command,omitempty"`
	With      map[string]Arg `json:"with,omitempty"`
	OnSuccess string         `json:"onSuccess,omitempty"`
}
