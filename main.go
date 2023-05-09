package main

import (
	"fmt"
	"os"

	"github.com/pomdtr/sunbeam/cmd"
)

var version = "dev"

func main() {
	rootCmd, err := cmd.NewCmdRoot(version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create root command: %s\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
