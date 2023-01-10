package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sunbeamlauncher/sunbeam/utils"
)

func NewCmdOpen() *cobra.Command {
	return &cobra.Command{
		Use:     "open",
		Short:   "Open file or url with default app",
		GroupID: "core",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return utils.Open(args[0])
		},
	}
}
