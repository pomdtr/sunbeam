package schemas

import (
	"embed"
	"encoding/json"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed *.schema.json
var embedFS embed.FS

var schemas map[string]*jsonschema.Schema

var schemaUrls = []string{
	"command.schema.json",
	"action.schema.json",
	"list.schema.json",
	"detail.schema.json",
	"form.schema.json",
	"page.schema.json",
	"manifest.schema.json",
}

func init() {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft7

	for _, url := range schemaUrls {
		schema, err := embedFS.Open(url)
		if err != nil {
			panic(err)
		}

		if err := compiler.AddResource(url, schema); err != nil {
			panic(err)
		}
	}

	schemas = make(map[string]*jsonschema.Schema)
	for _, url := range schemaUrls {
		schema, err := compiler.Compile(url)
		if err != nil {
			panic(err)
		}

		schemas[url] = schema
	}
}

func ValidatePage(input []byte) error {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	if err := schemas["page.schema.json"].Validate(v); err != nil {
		return err
	}
	return nil
}

func ValidateCommand(input []byte) error {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	if err := schemas["command.schema.json"].Validate(v); err != nil {
		return err
	}
	return nil
}

func ValidateManifest(input []byte) error {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	if err := schemas["manifest.schema.json"].Validate(v); err != nil {
		return err
	}
	return nil
}
