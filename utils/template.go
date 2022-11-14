package utils

import (
	"bytes"
	"text/template"

	"github.com/alessio/shellescape"
)

func RenderString(templateString string, inputs map[string]string) (string, error) {
	funcMap := make(template.FuncMap)
	for k, v := range inputs {
		funcMap[k] = func() string {
			return shellescape.Quote(v)
		}
	}
	t, err := template.New("").Funcs(funcMap).Parse(templateString)

	if err != nil {
		return "", err
	}

	out := bytes.Buffer{}
	if err = t.Execute(&out, nil); err != nil {
		return "", err
	}
	return out.String(), nil
}
