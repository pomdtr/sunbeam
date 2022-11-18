package cmd

import (
	"github.com/pomdtr/sunbeam/api"
)

type CustomFlag struct {
	input api.ScriptInput
	value string
}

func NewCustomFlag(input api.ScriptInput) *CustomFlag {
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
