package cmd

import (
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/spf13/cobra"
)

func NewCmdOpen() *cobra.Command {
	return &cobra.Command{
		Use:     "open [target]",
		GroupID: CommandGroupCore,
		Short:   "Open a file or url in your default application",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return utils.Open(args[0])
		},
	}
}
