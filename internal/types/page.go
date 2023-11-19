package types

type List struct {
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
	Actions []Action `json:"actions,omitempty"`
	Format  Format   `json:"format,omitempty"`
	Text    string   `json:"text,omitempty"`
}

type Format string

const (
	MarkdownFormat Format = "markdown"
	ANSIFormat     Format = "ansi"
	TemplateFormat Format = "template"
)
