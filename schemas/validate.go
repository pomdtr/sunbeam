package schemas

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed page.schema.json
var schemaString string
var schema = jsonschema.MustCompileString("", schemaString)

func Validate(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	if err := schema.Validate(v); err != nil {
		return fmt.Errorf("invalid schema: %#v", err)
	}

	return nil
}
