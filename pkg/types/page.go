package types

import "encoding/json"

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
	Title  string     `json:"title,omitempty"`
	Items  []ListItem `json:"items,omitempty"`
	Reload bool       `json:"reload,omitempty"`
}

func (l List) MarshalJSON() ([]byte, error) {
	type Alias List
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "list",
		Alias: (*Alias)(&l),
	})
}

type Detail struct {
	Title    string   `json:"title,omitempty"`
	Actions  []Action `json:"actions,omitempty"`
	Markdown string   `json:"markdown,omitempty"`
}

func (d Detail) MarshalJSON() ([]byte, error) {
	type Alias Detail
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "detail",
		Alias: (*Alias)(&d),
	})
}

type Form struct {
	Title string     `json:"title,omitempty"`
	Items []FormItem `json:"items,omitempty"`
}

func (f Form) MarshalJSON() ([]byte, error) {
	type Alias Form
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  "form",
		Alias: (*Alias)(&f),
	})
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
