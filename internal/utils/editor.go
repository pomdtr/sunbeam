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

func FindShell() string {
	if shell, ok := os.LookupEnv("SHELL"); ok {
		return shell
	}

	return "/bin/sh"
}

func FindPager() string {
	if pager, ok := os.LookupEnv("PAGER"); ok {
		return pager
	}

	return "less"
}
