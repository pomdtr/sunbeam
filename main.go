package main

import (
	"os"

	"github.com/pomdtr/sunbeam/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
