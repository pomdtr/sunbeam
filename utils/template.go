package utils

import (
	"bytes"
	"text/template"
)

func RenderString(s string, funcMap template.FuncMap) (string, error) {
	tmpl, err := template.New("template").Funcs(funcMap).Parse(s)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", err
	}

	return buf.String(), nil
}
