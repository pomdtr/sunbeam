package main

import (
	"fmt"
	"os"

	"github.com/sunbeamlauncher/sunbeam/cmd"
)

const version = "dev"

func main() {
	err := cmd.Execute(version)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
