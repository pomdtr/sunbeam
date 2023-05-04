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
		if ve, ok := err.(*jsonschema.ValidationError); ok {
			leaf := ve
			for len(leaf.Causes) > 0 {
				leaf = leaf.Causes[0]
			}
			return fmt.Errorf("%s does not validate with sunbeam schema", leaf.InstanceLocation)
		} else {
			return fmt.Errorf("invalid schema: %s", err)
		}
	}

	return nil
}
