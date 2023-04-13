package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	return &cobra.Command{
		Use:                "run <page>",
		Short:              "Run page from file",
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(args[0]); err == nil {
				command := exec.Command(args[0], args[1:]...)
				return Draw(command.Output)
			}

			if _, err := exec.LookPath(args[0]); err == nil {
				command := exec.Command(args[0], args[1:]...)
				return Draw(command.Output)
			}

			return fmt.Errorf("file or command not found: %s", args[0])
		},
	}

}
