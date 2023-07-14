package main

import (
	"os"

	"github.com/pomdtr/sunbeam/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
