package schemas

import (
	_ "embed"
	"encoding/json"

	"github.com/MakeNowJust/heredoc"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed page.schema.json
var schemaString string
var schema = jsonschema.MustCompileString("", schemaString)

type ValidationError struct {
	err error
	v   any
}

func (ve ValidationError) Error() string {
	return ve.err.Error()
}

func (ve ValidationError) Unwrap() error {
	return ve.err
}

func (ve ValidationError) Message() string {
	bs, err := json.MarshalIndent(ve.v, "", "  ")
	if err != nil {
		return err.Error()
	}

	codeblock := "```json\n" + string(bs) + "\n```"
	return heredoc.Docf(`
			## Page validation error

			%s

			Current page:

			%s`, ve, codeblock)
}

func Validate(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	if err := schema.Validate(v); err != nil {
		return ValidationError{
			err: err,
			v:   v,
		}
	}

	return nil
}
