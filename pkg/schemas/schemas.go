package schemas

import (
	_ "embed"
	"encoding/json"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed config.schema.json
var configSchemaString string
var ConfigSchema = jsonschema.MustCompileString("config.schema.json", configSchemaString)

//go:embed page.schema.json
var pageSchemaString string
var PageSchema = jsonschema.MustCompileString("page.schema.json", pageSchemaString)

//go:embed manifest.schema.json
var manifestSchemaString string
var ManifestSchema = jsonschema.MustCompileString("manifest.schema.json", manifestSchemaString)

func ValidatePage(input []byte) error {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	if err := PageSchema.Validate(v); err != nil {
		return err
	}
	return nil
}

func ValidateManifest(input []byte) error {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	if err := ManifestSchema.Validate(v); err != nil {
		return err
	}
	return nil
}

func ValidateConfig(input []byte) error {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	if err := ConfigSchema.Validate(v); err != nil {
		return err
	}
	return nil
}
