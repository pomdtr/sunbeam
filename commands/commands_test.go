package commands

import (
	"path"
	"testing"
)

func TestScanDir(t *testing.T) {
	scriptPath := path.Join(CommandDir, "github", "list-prs")
	_, err := Parse(scriptPath)

	if err != nil {
		t.Fatal(err)
	}
}
