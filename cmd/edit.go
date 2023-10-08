package cmd

import (
	"os"

	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/spf13/cobra"
)

func NewCmdEdit() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit a note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			editCmd := utils.EditCmd(args[0])
			editCmd.Stdin = os.Stdin
			editCmd.Stdout = os.Stdout
			editCmd.Stderr = os.Stderr

			return editCmd.Run()
		},
	}
}
