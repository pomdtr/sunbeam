package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Run a sunbeam command defined in the current directory",
		GroupID: coreGroupID,
	}

	if cwdCommand == nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("no command manifest found in current directory")
		}

		return cmd
	}

	if cwdCommand.Entrypoint != "" {
		cmd.DisableFlagParsing = true
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			var input string
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				input = string(b)
			}

			return runCommand(filepath.Join(cwd, cwdCommand.Entrypoint), args, input)
		}
		return cmd
	}

	for name, command := range cwdCommand.SubCommands {
		command := command
		cmd.AddCommand(&cobra.Command{
			Use:                name,
			Short:              command.Title,
			Long:               command.Description,
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				var input string
				if !isatty.IsTerminal(os.Stdin.Fd()) {
					b, err := io.ReadAll(os.Stdin)
					if err != nil {
						return err
					}
					input = string(b)
				}

				return runCommand(filepath.Join(cwd, command.Entrypoint), args, input)
			},
		})
	}

	return cmd
}
