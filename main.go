package main

import (
	"os"

	"github.com/SunbeamLauncher/sunbeam/cmd"
)

const version = "dev"

func main() {
	err := cmd.Execute(version)
	if err != nil {
		os.Exit(1)
	}
}
