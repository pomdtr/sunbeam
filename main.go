package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pomdtr/sunbeam/cmd"
)

func main() {
	if _, err := exec.LookPath("sunbeam"); err != nil {
		fmt.Fprintln(os.Stderr, "Please make sure that sunbeam is available in your $PATH")
		os.Exit(1)
	}

	rootCmd, err := cmd.NewRootCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
