package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <origin> [args...]",
		Short:              "Run an extension without installing it",
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveDefault
			}

			if len(args) > 1 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			if strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://") || args[0] == "-" {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			extension, err := tui.LoadExtension(args[0])
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			completions := make([]string, 0)
			for _, command := range extension.Commands {
				if command.Name == extension.Root {
					continue
				}
				completions = append(completions, fmt.Sprintf("%s\t%s", command.Name, command.Title))
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "--help" || args[0] == "-h" {
				return cmd.Help()
			}

			var scriptPath string
			if strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://") {
				res, err := http.Get(args[0])
				if err != nil {
					return err
				}
				defer res.Body.Close()

				if res.StatusCode != 200 {
					return fmt.Errorf("extension %s not found", args[0])
				}

				tempfile, err := os.CreateTemp("", "sunbeam-run-*.sh")
				if err != nil {
					return err
				}
				defer os.Remove(tempfile.Name())

				if _, err := io.Copy(tempfile, res.Body); err != nil {
					return err
				}

				if err := os.Chmod(tempfile.Name(), 0755); err != nil {
					return err
				}

				scriptPath = tempfile.Name()
			} else if args[0] == "-" {
				tempfile, err := os.CreateTemp("", "sunbeam-run-*.sh")
				if err != nil {
					return err
				}
				defer os.Remove(tempfile.Name())

				if _, err := io.Copy(tempfile, os.Stdin); err != nil {
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

				if info, err := os.Stat(s); err != nil {
					return err
				} else if info.IsDir() {
					scriptPath = filepath.Join(s, "sunbeam-extension")
					if _, err := os.Stat(scriptPath); err != nil {
						return fmt.Errorf("no extension found at %s", args[0])
					}
				} else {
					scriptPath = s
				}
			}

			rootCmd, err := NewCmdCustom(scriptPath)
			if err != nil {
				return fmt.Errorf("error loading extension: %w", err)
			}

			rootCmd.Use = args[0]
			rootCmd.SetArgs(args[1:])
			return rootCmd.Execute()
		},
	}

	return cmd
}
