package scripts

import (
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"

	_ "embed"

	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
)

//go:embed page.json
var pageSchema string

var Schema *jsonschema.Schema

func init() {
	var err error

	compiler := jsonschema.NewCompiler()

	if err = compiler.AddResource("https://pomdtr.github.io/sunbeam/schemas/page.json", strings.NewReader(pageSchema)); err != nil {
		panic(err)
	}
	Schema, err = compiler.Compile("https://pomdtr.github.io/sunbeam/schemas/page.json")
	if err != nil {
		panic(err)
	}
}

type Response struct {
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
	Text     string   `json:"text,omitempty"`
	Language string   `json:"language,omitempty"`
	Command  string   `json:"command,omitempty"`
	Args     []string `json:"args,omitempty"`
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

	Target string `json:"target,omitempty"`

	Text string `json:"text,omitempty"`

	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
}
