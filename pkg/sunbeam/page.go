package sunbeam

type List struct {
	Items              []ListItem `json:"items,omitempty"`
	EmptyText          string     `json:"emptyText,omitempty"`
	ShowDetail         bool       `json:"showDetail,omitempty"`
	AutoRefreshSeconds int        `json:"autoRefreshSeconds,omitempty"`
	Actions            []Action   `json:"actions,omitempty"`
}

type ListItem struct {
	Id          string         `json:"id,omitempty"`
	Title       string         `json:"title"`
	Subtitle    string         `json:"subtitle,omitempty"`
	Detail      ListItemDetail `json:"detail,omitempty"`
	Accessories []string       `json:"accessories,omitempty"`
	Actions     []Action       `json:"actions,omitempty"`
}

type ListItemDetail struct {
	Markdown string `json:"markdown,omitempty"`
	Text     string `json:"text,omitempty"`
}

type Detail struct {
	Actions  []Action `json:"actions,omitempty"`
	Markdown string   `json:"markdown,omitempty"`
	Text     string   `json:"text,omitempty"`
}
