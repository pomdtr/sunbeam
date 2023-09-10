package pkg

type List struct {
	Title     string     `json:"title,omitempty"`
	EmptyView *EmptyView `json:"emptyView,omitempty"`
	Items     []ListItem `json:"items"`
}

type Detail struct {
	Title     string   `json:"title,omitempty"`
	Actions   []Action `json:"actions,omitempty"`
	Highlight string   `json:"highlight,omitempty"`
	Text      string   `json:"text"`
}

type Form struct {
	Title        string  `json:"title,omitempty"`
	SubmitAction *Action `json:"submitAction,omitempty"`
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

type Input struct {
	Name        string        `json:"name"`
	Type        FormInputType `json:"type"`
	Title       string        `json:"title"`
	Placeholder string        `json:"placeholder,omitempty"`
	Default     any           `json:"default,omitempty"`
	Optional    bool          `json:"optional,omitempty"`

	// Only for dropdown
	Items []DropDownItem `json:"items,omitempty"`

	// Only for checkbox
	Label             string `json:"label,omitempty"`
	TrueSubstitution  string `json:"trueSubstitution,omitempty"`
	FalseSubstitution string `json:"falseSubstitution,omitempty"`
}

func NewTextInput(name string, title string, placeholder string) Input {
	return Input{
		Name:        name,
		Type:        TextInput,
		Title:       title,
		Placeholder: placeholder,
	}
}

func NewTextAreaInput(name string, title string, placeholder string) Input {
	return Input{
		Name:        name,
		Type:        TextAreaInput,
		Title:       title,
		Placeholder: placeholder,
	}
}

func NewCheckbox(name string, title string, label string) Input {
	return Input{
		Name:  name,
		Type:  CheckboxInput,
		Title: title,
		Label: label,
	}
}

func NewDropDown(name string, title string, items ...DropDownItem) Input {
	return Input{
		Name:  name,
		Type:  SelectInput,
		Title: title,
		Items: items,
	}
}

type ActionType string

const (
	ActionTypeCopy = "copy"
	ActionTypeOpen = "open"
	ActionTypeRun  = "run"
)

type Action struct {
	Title string     `json:"title,omitempty"`
	Type  ActionType `json:"type"`
	Key   string     `json:"key,omitempty"`

	// copy
	Text string `json:"text,omitempty"`

	// open
	Url string `json:"url,omitempty"`

	// run
	Command string         `json:"command,omitempty"`
	Params  map[string]any `json:"params,omitempty"`
}
