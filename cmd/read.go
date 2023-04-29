package cmd

import (
	"github.com/pomdtr/sunbeam/internal"
	"github.com/spf13/cobra"
)

func NewReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "read",
		Short:   "Read a file and push it",
		GroupID: coreGroupID,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return Draw(internal.NewFileGenerator(args[0]))
		},
	}
}
