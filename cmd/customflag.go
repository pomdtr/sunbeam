package cmd

import (
	"fmt"

	"github.com/pomdtr/sunbeam/app"
)

type CustomFlag struct {
	input app.ScriptParams
	value string
}

func NewCustomFlag(input app.ScriptParams) *CustomFlag {
	var value string
	switch input.Type {
	case "checkbox":
		value = "false"
	default:
		value = ""
	}
	return &CustomFlag{
		input: input,
		value: value,
	}
}

func (f *CustomFlag) String() string {
	return f.value
}

func (f *CustomFlag) Set(value string) error {
	if f.input.Type == "dropdown" {
		found := false
		for _, item := range f.input.Data {
			if item.Value == value {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid value for %s: %s", f.input.Name, value)
		}
	}
	f.value = value
	return nil
}

func (f *CustomFlag) Type() string {
	switch f.input.Type {
	case "textfield":
		return "string"
	case "textarea":
		return "string"
	case "dropdown":
		return "string"
	case "checkbox":
		return "bool"
	default:
		return "string"
	}
}
