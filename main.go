package main

import (
	"os"

	"github.com/pomdtr/sunbeam/cmd"
)

var version = "dev"

func main() {
	err := cmd.Execute(version)
	if err != nil {
		os.Exit(1)
	}
}
