package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewRequireCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "require",
		Short: "Check if command is installed",
		Args:  cobra.ArbitraryArgs,
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
