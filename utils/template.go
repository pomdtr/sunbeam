package utils

import (
	"bytes"
	"text/template"
)

func RenderString(templateString string, funcMap template.FuncMap) (string, error) {

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
