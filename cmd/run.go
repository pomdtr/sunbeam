package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewCmdRun(extensionDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run",
		Short:              "Generate a page from a command or a script, and push it's output",
		GroupID:            coreGroupID,
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input string
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				input = string(b)
			}

			if strings.HasPrefix(args[0], ".") {
				if _, err := os.Stat(args[0]); err != nil {
					return fmt.Errorf("%s: no such file or directory", args[0])
				}

				return Run(internal.NewCommandGenerator(&types.Command{
					Name:  args[0],
					Args:  args[1:],
					Input: input,
				}))
			}

			if _, err := exec.LookPath(args[0]); err == nil {
				return Run(internal.NewCommandGenerator(&types.Command{
					Name:  args[0],
					Args:  args[1:],
					Input: input,
				}))
			}

			return fmt.Errorf("file or command not found: %s", args[0])
		},
	}

	return cmd
}
