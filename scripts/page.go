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
	Type    string   `json:"type"`
	Title   string   `json:"title"`
	Actions []Action `json:"actions"`

	*Detail
	*List
}

type Detail struct {
	Content Preview `json:"content"`
}

type List struct {
	ShowPreview   bool       `json:"showPreview"`
	GenerateItems bool       `json:"generateItems"`
	Items         []ListItem `json:"items"`
}

type Preview struct {
	Text     string   `json:"text"`
	Language string   `json:"language"`
	Command  []string `json:"command"`
}

type ListItem struct {
	Id          string   `json:"id"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Preview     *Preview `json:"preview"`
	Accessories []string `json:"accessories"`
	Actions     []Action `json:"actions"`
}

type Action struct {
	Title string `json:"title"`
	Type  string `json:"type"`

	Shortcut string `json:"shortcut"`

	Target string `json:"target"`

	Text string `json:"text"`

	Command []string `json:"command"`
}
