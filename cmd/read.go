package cmd

import (
	"github.com/pomdtr/sunbeam/internal"
	"github.com/spf13/cobra"
)

func NewReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <page>",
		Short: "Read page from file, and push it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return Draw(internal.NewFileGenerator(args[0]))
		},
	}

	return cmd
}
