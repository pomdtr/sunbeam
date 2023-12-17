package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func RunCommand(cmd string, dir string) error {
	// Parse the command
	file, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	if err != nil {
		fmt.Printf("Error parsing command: %v\n", err)
	}

	// Create a shell interpreter
	sh, err := interp.New(interp.StdIO(os.Stdin, os.Stdout, os.Stderr), interp.Dir(dir))
	if err != nil {
		fmt.Printf("Error creating shell: %v\n", err)
	}

	// Create a shell interpreter
	return sh.Run(context.Background(), file)
}
