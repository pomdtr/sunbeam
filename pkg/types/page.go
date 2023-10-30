package types

type List struct {
	Title     string     `json:"title,omitempty"`
	Items     []ListItem `json:"items,omitempty"`
	Dynamic   bool       `json:"dynamic,omitempty"`
	EmptyText string     `json:"emptyText,omitempty"`
	Actions   []Action   `json:"actions,omitempty"`
}

type ListItem struct {
	Id          string   `json:"id,omitempty"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle,omitempty"`
	Accessories []string `json:"accessories,omitempty"`
	Actions     []Action `json:"actions,omitempty"`
}

type Detail struct {
	Title     string    `json:"title,omitempty"`
	Actions   []Action  `json:"actions,omitempty"`
	Highlight Highlight `json:"highlight,omitempty"`
	Text      string    `json:"text,omitempty"`
}

type Highlight string

const (
	HighlightMarkdown Highlight = "markdown"
	HighlightAnsi     Highlight = "ansi"
)
