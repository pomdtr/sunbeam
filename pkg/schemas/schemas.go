package schemas

import (
	"embed"
	"encoding/json"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed *.schema.json
var embedFS embed.FS

var schemas map[string]*jsonschema.Schema

const (
	CommandSchemaURL  = "command.schema.json"
	InputSchemaURL    = "input.schema.json"
	PageSchemaURL     = "page.schema.json"
	ManifestSchemaURL = "manifest.schema.json"
)

var schemaUrls = []string{
	CommandSchemaURL,
	InputSchemaURL,
	PageSchemaURL,
	ManifestSchemaURL,
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

	if err := schemas[PageSchemaURL].Validate(v); err != nil {
		return err
	}
	return nil
}

func ValidateCommand(input []byte) error {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	if err := schemas[CommandSchemaURL].Validate(v); err != nil {
		return err
	}
	return nil
}

func ValidateManifest(input []byte) error {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	if err := schemas[ManifestSchemaURL].Validate(v); err != nil {
		return err
	}
	return nil
}
