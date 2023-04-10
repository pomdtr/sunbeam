package schemas

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed page.schema.json
var schemaString string
var schema = jsonschema.MustCompileString("", schemaString)

type PrettyValidationError struct {
	*jsonschema.ValidationError
}

func NewPrettyValidationError(err *jsonschema.ValidationError) *PrettyValidationError {
	return &PrettyValidationError{err}
}

func prettifyValidationError(err *jsonschema.ValidationError, prefix string) string {
	var msg string
	if err.InstanceLocation == "" {
		msg = fmt.Sprintf("%s%s", prefix, err.Message)
	} else {
		msg = fmt.Sprintf("%s%s: %s", prefix, err.InstanceLocation, err.Message)
	}
	for _, c := range err.Causes {
		for _, line := range strings.Split(prettifyValidationError(c, "L "), "\n") {
			msg += "\n  " + line
		}
	}

	return msg
}

func (e *PrettyValidationError) Error() string {

	return prettifyValidationError(e.ValidationError, "")
}

func Validate(v any) error {
	if err := schema.Validate(v); err != nil {
		if ve, ok := err.(*jsonschema.ValidationError); ok {
			return NewPrettyValidationError(ve)
		}
		return err
	}

	return nil
}
