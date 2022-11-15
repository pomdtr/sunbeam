package utils

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/alessio/shellescape"
)

func RenderString(templateString string, inputs map[string]string) (string, error) {
	funcMap := template.FuncMap{}
	funcMap["input"] = func(input string) (string, error) {
		if value, ok := inputs[input]; ok {
			return shellescape.Quote(value), nil
		}
		return "", fmt.Errorf("input %s not found", input)
	}

	t, err := template.New("").Delims("${{", "}}").Funcs(funcMap).Parse(templateString)

	if err != nil {
		return "", err
	}

	out := bytes.Buffer{}
	if err = t.Execute(&out, nil); err != nil {
		return "", err
	}
	return out.String(), nil
}
