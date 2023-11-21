package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run",
		Short:              "Run a command",
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(1),

		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveDefault
			}

			entrypoint, err := filepath.Abs(args[0])
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			extension, err := extensions.ExtractManifest(entrypoint)
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			var completions []string
			for _, command := range extension.Commands {
				completions = append(completions, fmt.Sprintf("%s\t%s", command.Name, command.Title))
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var entrypoint string
			if args[0] == "-" {
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

				entrypoint = tempfile.Name()
			} else if extensions.IsRemote(args[0]) {
				tempfile, err := os.CreateTemp("", "entrypoint-*%s")
				if err != nil {
					return err
				}
				defer os.Remove(tempfile.Name())

				if err := extensions.DownloadEntrypoint(args[0], tempfile.Name()); err != nil {
					return err
				}

				entrypoint = tempfile.Name()
			} else {
				e, err := filepath.Abs(args[0])
				if err != nil {
					return err
				}

				if _, err := os.Stat(e); err != nil {
					return fmt.Errorf("error loading extension: %w", err)
				}

				entrypoint = e
			}

			if err := os.Chmod(entrypoint, 0755); err != nil {
				return err
			}

			manifest, err := extensions.ExtractManifest(entrypoint)
			if err != nil {
				return fmt.Errorf("error loading extension: %w", err)
			}

			rootCmd, err := NewCmdCustom(filepath.Base(entrypoint), extensions.Extension{
				Manifest:   manifest,
				Entrypoint: entrypoint,
			}, extensions.Config{
				Origin: entrypoint,
			})
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
