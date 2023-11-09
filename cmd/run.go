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
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveDefault
			}

			if len(args) == 1 {
				entrypoint, err := filepath.Abs(args[0])
				if err != nil {
					return nil, cobra.ShellCompDirectiveNoFileComp
				}

				extension, err := ExtractManifest(entrypoint)
				if err != nil {
					return nil, cobra.ShellCompDirectiveNoFileComp
				}

				var commands []string
				for _, command := range extension.Commands {
					commands = append(commands, command.Name)
				}

				return commands, cobra.ShellCompDirectiveNoFileComp
			}

			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		GroupID: CommandGroupCore,
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

				scriptPath = tempfile.Name()
			} else {
				s, err := filepath.Abs(args[0])
				if err != nil {
					return err
				}

				if _, err := os.Stat(s); err != nil {
					return fmt.Errorf("error loading extension: %w", err)
				}

				scriptPath = s
			}

			extension, err := ExtractManifest(scriptPath)
			if err != nil {
				return fmt.Errorf("error loading extension: %w", err)
			}

			rootCmd, err := NewCmdCustom(filepath.Base(scriptPath), extension, nil)
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
