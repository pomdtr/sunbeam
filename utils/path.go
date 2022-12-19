package utils

import (
	"path"
)

func IsRoot(filepath string) bool {
	return path.Dir(filepath) == filepath
}
