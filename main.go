package main

import (
	"fmt"
	"os"

	"github.com/pomdtr/sunbeam/cmd"
)

func init() {
	if _, ok := os.LookupEnv("CSB"); ok {
		_ = os.Setenv("CSB", "1")
	}
}

func main() {
	rootCmd, err := cmd.NewRootCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
