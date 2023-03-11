package schemas

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"

	_ "embed"

	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
)

//go:embed schema.json
var schemaString string

var Schema *jsonschema.Schema

func init() {
	var err error

	compiler := jsonschema.NewCompiler()

	if err = compiler.AddResource("https://pomdtr.github.io/sunbeam/schemas/page.json", strings.NewReader(schemaString)); err != nil {
		panic(err)
	}
	Schema, err = compiler.Compile("https://pomdtr.github.io/sunbeam/schemas/page.json")
	if err != nil {
		panic(err)
	}
}

func Validate(bytes []byte) error {
	var v any
	if err := json.Unmarshal(bytes, &v); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	return Schema.Validate(v)
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
	Command  string   `json:"command"`
	Args     []string `json:"args"`
}

type ListItem struct {
	Id          string   `json:"id"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Preview     *Preview `json:"preview"`
	Accessories []string `json:"accessories"`
	Actions     []Action `json:"actions"`
}

type FormInput struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Placeholder string `json:"placeholder"`
	Default     string `json:"default"`

	Choices []string `json:"choices"`
	Label   string   `json:"label"`
}

type Action struct {
	Title    string `json:"title"`
	Shortcut string `json:"shortcut"`
	Type     string `json:"type"`

	Inputs []FormInput `json:"inputs"`

	// edit
	Page string `json:"page"`

	// open
	Target string `json:"target"`

	// copy
	Text string `json:"text"`

	// run / push
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Dir     string

	// run
	OnSuccess string `json:"onSuccess"`
}
