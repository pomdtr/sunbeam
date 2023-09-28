package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var template = `#/bin/sh

exec sunbeam fetch %s "$@"
`

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <origin> [args...]",
		Short:              "Run an extension without installing it",
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "--help" || args[0] == "-h" {
				return cmd.Help()
			}

			var scriptPath string
			if strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://") {
				tempfile, err := os.CreateTemp("", "sunbeam-run-*.sh")
				if err != nil {
					return err
				}
				defer os.Remove(tempfile.Name())

				if err := os.Chmod(tempfile.Name(), 0755); err != nil {
					return err
				}

				if _, err := tempfile.WriteString(fmt.Sprintf(template, args[0])); err != nil {
					return err
				}

				scriptPath = tempfile.Name()
			} else {
				scriptPath = args[0]
			}

			rootCmd, err := NewExtensionCommand(scriptPath)
			if err != nil {
				return err
			}

			rootCmd.Use = args[0]
			rootCmd.SetArgs(args[1:])
			return rootCmd.Execute()
		},
	}

	return cmd
}
