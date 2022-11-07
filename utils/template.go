package utils

import (
	"bytes"
	"html/template"
	"os"
	"strings"
)

var envMap map[string]string

func init() {
	envMap = make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		envMap[parts[0]] = parts[1]
	}
}

func RenderString(templateString string, inputs map[string]string) (string, error) {
	t, err := template.New("").Delims("${", "}").Funcs(template.FuncMap{
		"env": func(key string) string {
			env, ok := envMap[key]
			if !ok {
				return ""
			}
			return env
		},
		"input": func(key string) string {
			input, ok := inputs[key]
			if !ok {
				return ""
			}
			return input
		},
	}).Parse(templateString)

	if err != nil {
		return "", err
	}

	out := bytes.Buffer{}
	if err = t.Execute(&out, nil); err != nil {
		return "", err
	}
	return out.String(), nil
}
