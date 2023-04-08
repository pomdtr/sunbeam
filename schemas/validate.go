package schemas

import (
	_ "embed"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed page.schema.json
var schemaString string
var schema = jsonschema.MustCompileString("", schemaString)

func Validate(v any) error {
	return schema.Validate(v)
}
