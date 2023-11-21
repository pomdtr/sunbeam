package main

import (
	"fmt"
	"os"

	"github.com/pomdtr/sunbeam/internal/cli"
)

func main() {
	rootCmd, err := cli.NewRootCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
