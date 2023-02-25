package utils

import (
	"os"
	"path"
	"strings"
)

func CopyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

func ResolvePath(filepath string) (string, error) {
	if strings.HasPrefix(filepath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return strings.Replace(filepath, "~", homeDir, 1), nil
	}

	if !path.IsAbs(filepath) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return path.Join(cwd, filepath), nil
	}

	return filepath, nil

}

func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)

	return !os.IsNotExist(err)
}
