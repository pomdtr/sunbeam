package types

type Page struct {
	Type  PageType `json:"type"`
	Title string   `json:"title,omitempty"`
}

type PageType string

const (
	PageTypeList   PageType = "list"
	PageTypeForm   PageType = "form"
	PageTypeDetail PageType = "detail"
)

type List struct {
	Items []ListItem `json:"items"`
}

type Detail struct {
	Actions  []Action `json:"actions,omitempty"`
	Text     string   `json:"text,omitempty"`
	Language string   `json:"language,omitempty"`
}

type Form struct {
	Items []FormItem `json:"items,omitempty"`
}

type ListItem struct {
	Id          string   `json:"id,omitempty"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle,omitempty"`
	Accessories []string `json:"accessories,omitempty"`
	Actions     []Action `json:"actions,omitempty"`
}

type Metadata struct {
	Title  string `json:"title,omitempty"`
	Text   string `json:"text,omitempty"`
	Target string `json:"target,omitempty"`
}

type FormItemType string

const (
	TextInput     FormItemType = "text"
	TextAreaInput FormItemType = "textarea"
	SelectInput   FormItemType = "select"
	CheckboxInput FormItemType = "checkbox"
)

type DropDownItem struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type FormItem struct {
	Type        FormItemType `json:"type"`
	Name        string       `json:"name,omitempty"`
	Title       string       `json:"title"`
	Placeholder string       `json:"placeholder,omitempty"`
	Default     any          `json:"default,omitempty"`
	Optional    bool         `json:"optional,omitempty"`

	// Only for dropdown
	Items []DropDownItem `json:"items,omitempty"`

	// Only for checkbox
	Label string `json:"label,omitempty"`
}

type Action struct {
	Title    string  `json:"title,omitempty"`
	Key      string  `json:"key,omitempty"`
	OnAction Command `json:"onAction,omitempty"`
}
