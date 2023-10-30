package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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
				origin, err := url.Parse(args[0])
				if err != nil {
					return err
				}
				pattern := fmt.Sprintf("entrypoint-*%s", filepath.Ext(origin.Path))

				tempfile, err := os.CreateTemp("", pattern)
				if err != nil {
					return err
				}
				defer os.Remove(tempfile.Name())

				resp, err := http.Get(args[0])
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("error downloading extension: %s", resp.Status)
				}

				if _, err := io.Copy(tempfile, resp.Body); err != nil {
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
				tempfile, err := os.CreateTemp("", "entrypoint-*%s")
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

				if _, err := os.Stat(s); err != nil {
					return fmt.Errorf("error loading extension: %w", err)
				}

				if err := os.Chmod(s, 0755); err != nil {
					return err
				}

				scriptPath = s
			}

			alias := filepath.Base(scriptPath)
			extension, err := LoadExtension(scriptPath)
			extension.Alias = alias
			if err != nil {
				return err
			}

			rootCmd, err := NewCmdCustom(extension)
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
