package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewCmdWrap() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "wrap",
		Short:              "Wrap a command to be used interactively from a sunbeam extension",
		Hidden:             true,
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wrapped := exec.Command(args[0], args[1:]...)

			f, err := os.Open("/dev/tty")
			if err != nil {
				return err
			}

			wrapped.Stdin = f
			wrapped.Stdout = os.Stderr

			return wrapped.Run()
		},
	}

	return cmd
}
