package utils

import (
	"encoding/json"
	"os"
	"path"
	"strings"
)

func ReadJson(filepath string, v interface{}) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(v)
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

func IsRoot(filepath string) bool {
	return path.Dir(filepath) == filepath
}
