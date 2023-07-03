package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pomdtr/sunbeam/cmd"
)

var (
	dataDir      = filepath.Join(xdg.DataHome, "sunbeam")
	manifestPath = filepath.Join(dataDir, "commands.json")
)

func main() {
	manifest, err := cmd.LoadManifest(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := cmd.NewRootCmd(manifest).Execute(); err != nil {
		os.Exit(1)
	}
}
