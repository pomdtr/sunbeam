package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

type ExitCodeError struct {
	ExitCode int
}

func (e ExitCodeError) Error() string {
	return fmt.Sprintf("exit code %d", e.ExitCode)
}

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run <command> [args...]",
		Short:   "read a sunbeam command",
		GroupID: coreGroupID,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input string
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				bs, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				input = string(bs)
			}

			info, err := os.Stat(args[0])
			if err != nil {
				return err
			}

			scriptPath := args[0]
			if info.IsDir() {
				root, err := findRoot(scriptPath)
				if err != nil {
					return err
				}

				scriptPath = root
			}

			if filepath.Base(scriptPath) != manifestName {
				return runCommand(types.Command{
					Name:  scriptPath,
					Args:  args[1:],
					Input: input,
				})
			}

			command, err := ParseCommand(filepath.Dir(scriptPath))
			if err != nil {
				return fmt.Errorf("could not parse command: %w", err)
			}

			if command.Entrypoint != "" {
				return runCommand(types.Command{
					Name:  filepath.Join(filepath.Dir(scriptPath), command.Entrypoint),
					Args:  args[1:],
					Input: input,
				})
			}

			if len(args) < 2 {
				cmd.PrintErrln("No subcommand provided")
				cmd.PrintErrln("Subcommands:")
				for name := range command.SubCommands {
					cmd.PrintErrf("  - %s\n", name)
				}
				return ExitCodeError{ExitCode: 1}
			}

			subcommand := args[1]
			for name, c := range command.SubCommands {
				if subcommand != name {
					continue
				}

				return runCommand(types.Command{
					Name:  filepath.Join(filepath.Dir(scriptPath), c.Entrypoint),
					Args:  args[2:],
					Input: input,
				})
			}

			return fmt.Errorf("subcommand not found: %s", subcommand)
		},
	}

	return cmd
}
