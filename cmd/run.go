package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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
		Use:     "run <origin> [args...]",
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

			if originUrl, err := url.Parse(args[0]); err == nil && originUrl.Scheme == "https" {
				resp, err := http.Get(args[0])
				if err != nil {
					return fmt.Errorf("could not fetch script: %w", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != 200 {
					return fmt.Errorf("could not fetch script: %s", resp.Status)
				}

				tmpDir, err := os.MkdirTemp("", "sunbeam")
				if err != nil {
					return err
				}
				defer os.RemoveAll(tmpDir)

				scriptPath := filepath.Join(tmpDir, "sunbeam-command")
				f, err := os.OpenFile(scriptPath, os.O_CREATE|os.O_WRONLY, 0755)
				if err != nil {
					return err
				}
				defer f.Close()

				if _, err := io.Copy(f, resp.Body); err != nil {
					return err
				}
				if err := f.Close(); err != nil {
					return err
				}

				return runCommand(types.Command{
					Name:  scriptPath,
					Args:  args[1:],
					Input: input,
				})
			}

			scriptPath := args[0]
			if info, err := os.Stat(scriptPath); err != nil {
				return fmt.Errorf("could not find script: %w", err)
			} else if info.IsDir() {
				root, err := findRoot(scriptPath)
				if err != nil {
					return err
				}

				scriptPath = root
			}

			if filepath.Base(scriptPath) != "sunbeam.json" {
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

			if len(command.SubCommands) == 0 {
				return runCommand(types.Command{
					Name:  filepath.Join(filepath.Dir(scriptPath), command.Command),
					Args:  args[1:],
					Input: input,
				})
			}

			if len(args) == 1 {
				if command.Command != "" {
					return runCommand(types.Command{
						Name:  filepath.Join(filepath.Dir(scriptPath), command.Command),
						Args:  args[1:],
						Input: input,
					})
				}

				cmd.PrintErrln("No subcommand provided!")
				cmd.PrintErrln()
				cmd.PrintErrln("Available subcommands:")
				for name := range command.SubCommands {
					cmd.PrintErrf(" - %s\n", name)
				}
				cmd.PrintErrln()
				return ExitCodeError{ExitCode: 1}
			}

			subcommand, ok := command.SubCommands[args[1]]
			if !ok {
				if command.Command != "" {
					return runCommand(types.Command{
						Name:  filepath.Join(filepath.Dir(scriptPath), command.Command),
						Args:  args[1:],
						Input: input,
					})
				}

				cmd.PrintErrf("Subcommand not found: %s\n!", args[1])
				cmd.PrintErrln()
				cmd.PrintErrln("Subcommands:")
				for name := range command.SubCommands {
					cmd.PrintErrf("  - %s\n", name)
				}
				cmd.PrintErrln()
				return ExitCodeError{ExitCode: 1}
			}

			return runCommand(types.Command{
				Name:  filepath.Join(filepath.Dir(scriptPath), subcommand.Command),
				Args:  args[2:],
				Input: input,
			})

		},
	}

	return cmd
}
