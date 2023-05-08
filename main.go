package main

import (
	"fmt"
	"os"

	"github.com/pomdtr/sunbeam/cmd"
)

var version = "dev"

func main() {
	if err := cmd.Execute(version); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
