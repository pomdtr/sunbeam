package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewRequireCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "require <command> [command...]",
		Short: "Check if command is installed",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if _, err := exec.LookPath(arg); err != nil {
					return fmt.Errorf("command not found: %s", arg)
				}
			}

			return nil
		},
	}
}
