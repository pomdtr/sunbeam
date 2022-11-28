package cmd

import (
	"fmt"
	"strconv"

	"github.com/pomdtr/sunbeam/app"
)

type CustomFlag struct {
	input app.ScriptParam
	value any
}

func NewCustomFlag(input app.ScriptParam) *CustomFlag {
	return &CustomFlag{
		input: input,
		value: input.Default,
	}
}

func (f *CustomFlag) String() string {
	return fmt.Sprintf("%v", f.value)
}

func (f *CustomFlag) Set(value string) error {
	var v any
	switch f.input.Type {
	case "string":
		v = value
	case "boolean":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v = b
	default:
		return fmt.Errorf("unknown type %s", f.input.Type)
	}

	if f.input.Enum != nil {
		for _, value := range f.input.Enum {
			if v == value {
				f.value = v
				return nil
			}
		}
		return fmt.Errorf("invalid value %v, must be one of %v", v, f.input.Enum)
	}

	f.value = value
	return nil
}

func (f *CustomFlag) Type() string {
	switch f.input.Type {
	case "string":
		return "string"
	case "boolean":
		return "bool"
	default:
		return "string"
	}
}
