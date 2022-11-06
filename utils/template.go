package utils

import (
	"bytes"
	"html/template"
)

func RenderString(templateString string, data map[string]any) (string, error) {
	t, err := template.New("").Parse(templateString)
	if err != nil {
		return "", err
	}

	out := bytes.Buffer{}
	if err = t.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}
