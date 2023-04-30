package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewCmdRun(extensionDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Generate a page from a command or a script, and push it",
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
			if args[0] == "." {
				if _, err := os.Stat(extensionBinaryName); err != nil {
					return fmt.Errorf("no extension found in current directory")
				}

				abs, err := filepath.Abs(extensionBinaryName)
				if err != nil {
					return fmt.Errorf("unable to get absolute path: %s", err)
				}

				return Draw(internal.NewCommandGenerator(&types.Command{
					Name: abs,
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

	cmd.Flags().String("on-success", "push", "action to trigger when the command is successful")

	return cmd
}
