package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewCmdRun(extensionDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Generate a page from a command or a script, and push it's output",
		GroupID: coreGroupID,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			extension, err := ListExtensions(extensionDir)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			return extension, cobra.ShellCompDirectiveDefault
		},
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			if args[0] == "." {
				if _, err := os.Stat(extensionBinaryName); err != nil {
					return fmt.Errorf("no extension found in current directory")
				}

				return runExtension(cwd, args)
			}

			if strings.HasPrefix(args[0], ".") {
				if _, err := os.Stat(args[0]); err != nil {
					return fmt.Errorf("%s: no such file or directory", args[0])
				}

				return Draw(internal.NewCommandGenerator(&types.Command{
					Name: args[0],
					Args: args[1:],
				}))
			}

			if _, err := exec.LookPath(args[0]); err == nil {
				return Draw(internal.NewCommandGenerator(&types.Command{
					Name: args[0],
					Args: args[1:],
				}))
			}

			return fmt.Errorf("file or command not found: %s", args[0])
		},
	}

	return cmd
}
