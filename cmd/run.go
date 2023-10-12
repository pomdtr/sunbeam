package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <origin> [args...]",
		Short:              "Run an extension from a script, directory, or URL",
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		GroupID:            CommandGroupCore,
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "--help" || args[0] == "-h" {
				return cmd.Help()
			}

			s, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			info, err := os.Stat(s)
			if err != nil {
				return fmt.Errorf("error loading extension: %w", err)
			}

			var scriptPath string
			if info.IsDir() {
				scriptPath = filepath.Join(s, "sunbeam-extension")
				if _, err := os.Stat(scriptPath); err != nil {
					return fmt.Errorf("no extension found at %s", args[0])
				}
			} else {
				scriptPath = s
			}

			extension, err := LoadExtension(scriptPath)
			if err != nil {
				return err
			}

			extensionMap, err := FindExtensions()
			if err != nil {
				return err
			}

			extensionMap[args[0]] = extension
			rootCmd, err := NewCmdCustom(extensionMap, args[0])
			if err != nil {
				return fmt.Errorf("error loading extension: %w", err)
			}

			rootCmd.Use = "extension"
			rootCmd.SetArgs(args[1:])
			return rootCmd.Execute()
		},
	}

	return cmd
}
