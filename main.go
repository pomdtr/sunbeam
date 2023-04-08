package main

import (
	"fmt"
	"os"

	_ "embed"

	"github.com/pomdtr/sunbeam/cmd"
)

var version = "dev"

func main() {
	cmd, err := cmd.NewRootCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create root command: %s", err)
		os.Exit(1)
	}

	cmd.Version = version
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
