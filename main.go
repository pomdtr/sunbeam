package main

import (
	"os"

	_ "embed"

	"github.com/pomdtr/sunbeam/cmd"
)

var version = "dev"

func main() {
	if err := cmd.Execute(version); err != nil {
		os.Exit(1)
	}
}
