package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

			var scriptPath string
			if strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://") {
				tempfile, err := os.CreateTemp("", "sunbeam-extension-")
				if err != nil {
					return err
				}
				defer os.Remove(tempfile.Name())

				if err := renderHTTPEntrypoint(args[0], tempfile); err != nil {
					return err
				}

				if err := tempfile.Close(); err != nil {
					return err
				}

				if err := os.Chmod(tempfile.Name(), 0755); err != nil {
					return err
				}

				scriptPath = tempfile.Name()
			} else if args[0] == "-" {
				tempfile, err := os.CreateTemp("", "sunbeam-extension-")
				if err != nil {
					return err
				}
				defer os.Remove(tempfile.Name())

				if _, err := io.Copy(tempfile, os.Stdin); err != nil {
					return err
				}

				if err := tempfile.Close(); err != nil {
					return err
				}

				if err := os.Chmod(tempfile.Name(), 0755); err != nil {
					return err
				}

				scriptPath = tempfile.Name()
			} else {
				s, err := filepath.Abs(args[0])
				if err != nil {
					return err
				}

				info, err := os.Stat(s)
				if err != nil {
					return fmt.Errorf("error loading extension: %w", err)
				}

				if info.IsDir() {
					scriptPath = filepath.Join(s, "sunbeam-extension")
					if _, err := os.Stat(scriptPath); err != nil {
						return fmt.Errorf("no extension found at %s", args[0])
					}
				} else {
					scriptPath = s
				}
			}

			extension, err := LoadExtension(scriptPath)
			if err != nil {
				return err
			}

			rootCmd, err := NewCmdCustom(args[0], extension)
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
