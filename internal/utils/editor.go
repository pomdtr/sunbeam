package utils

import "os"

func FindEditor() string {
	if editor, ok := os.LookupEnv("VISUAL"); ok {
		return editor
	}

	if editor, ok := os.LookupEnv("EDITOR"); ok {
		return editor
	}

	return "vi"
}
