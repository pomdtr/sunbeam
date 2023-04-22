package main

import (
	"os"

	_ "embed"

	"github.com/pomdtr/sunbeam/cmd"
)

var version = "dev"

func main() {
	cmd := cmd.NewRootCmd(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
