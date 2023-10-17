package types

import "encoding/json"

type View struct {
	Type  ViewType `json:"type"`
	Title string   `json:"title,omitempty"`
}

type ViewType string

const (
	ViewTypeList   ViewType = "list"
	ViewTypeDetail ViewType = "detail"
)

type List struct {
	Title     string     `json:"title,omitempty"`
	Items     []ListItem `json:"items,omitempty"`
	Dynamic   bool       `json:"dynamic,omitempty"`
	EmptyText string     `json:"emptyText,omitempty"`
	Actions   []Action   `json:"actions,omitempty"`
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

type ListItem struct {
	Id          string   `json:"id,omitempty"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle,omitempty"`
	Accessories []string `json:"accessories,omitempty"`
	Actions     []Action `json:"actions,omitempty"`
}

type Action struct {
	Title    string  `json:"title,omitempty"`
	Key      string  `json:"key,omitempty"`
	OnAction Command `json:"onAction,omitempty"`
}
