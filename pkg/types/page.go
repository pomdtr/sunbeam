package types

type PageType string

const (
	PageTypeList   PageType = "list"
	PageTypeForm   PageType = "form"
	PageTypeDetail PageType = "detail"
)

type List struct {
	Title     string     `json:"title,omitempty"`
	EmptyView *EmptyView `json:"emptyView,omitempty"`
	Items     []ListItem `json:"items"`
}

type Detail struct {
	Title    string   `json:"title,omitempty"`
	Actions  []Action `json:"actions,omitempty"`
	Text     string   `json:"text,omitempty"`
	Language string   `json:"language,omitempty"`
}

type Form struct {
	Title   string     `json:"title,omitempty"`
	Inputs  []FormItem `json:"inputs,omitempty"`
	Command CommandRef `json:"command,omitempty"`
}

type EmptyView struct {
	Text    string   `json:"text,omitempty"`
	Actions []Action `json:"actions,omitempty"`
}

type ListItem struct {
	Id          string   `json:"id,omitempty"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle,omitempty"`
	Accessories []string `json:"accessories,omitempty"`
	Actions     []Action `json:"actions,omitempty"`
}

type Metadata struct {
	Title string `json:"title,omitempty"`
	Text  string `json:"text,omitempty"`
	Url   string `json:"url,omitempty"`
}

type FormInputType string

const (
	TextInput     FormInputType = "text"
	TextAreaInput FormInputType = "textarea"
	SelectInput   FormInputType = "select"
	CheckboxInput FormInputType = "checkbox"
)

type DropDownItem struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type FormItem struct {
	Type        FormInputType `json:"type"`
	Name        string        `json:"name,omitempty"`
	Title       string        `json:"title"`
	Placeholder string        `json:"placeholder,omitempty"`
	Default     any           `json:"default,omitempty"`
	Optional    bool          `json:"optional,omitempty"`

	// Only for dropdown
	Items []DropDownItem `json:"items,omitempty"`

	// Only for checkbox
	Label string `json:"label,omitempty"`
}

type ActionType string

const (
	ActionTypeRun    ActionType = "run"
	ActionTypeOpen   ActionType = "open"
	ActionTypeCopy   ActionType = "copy"
	ActionTypeReload ActionType = "reload"
)

type Action struct {
	Title  string     `json:"title,omitempty"`
	Key    string     `json:"key,omitempty"`
	Type   ActionType `json:"type,omitempty"`
	Exit   bool       `json:"exit,omitempty"`
	Reload bool       `json:"reload,omitempty"`

	Text string `json:"text,omitempty"`

	Url string `json:"url,omitempty"`

	Command CommandRef `json:"command,omitempty"`
}

type CommandRef struct {
	Origin string         `json:"origin,omitempty"`
	Name   string         `json:"name"`
	Params map[string]any `json:"params,omitempty"`
}
