package main

import (
	"os"

	"github.com/pomdtr/sunbeam/cmd"
)

var version = "dev"

func main() {
	cmd := cmd.NewRootCmd(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
