package utils

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/alecthomas/chroma/quick"
)

func RenderString(templateString string, funcMap template.FuncMap) (string, error) {
	t, err := template.New("").Funcs(funcMap).Delims("{{", "}}").Parse(templateString)

	if err != nil {
		return "", err
	}

	out := bytes.Buffer{}
	if err = t.Execute(&out, nil); err != nil {
		return "", err
	}
	return out.String(), nil
}

func HighlightString(source string, lang string) (string, error) {
	if lang == "" {
		return source, nil
	}

	builder := strings.Builder{}
	err := quick.Highlight(&builder, source, lang, "terminal16", "github")
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}
