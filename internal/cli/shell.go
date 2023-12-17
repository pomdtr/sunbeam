package cli

import (
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/spf13/cobra"
)

func NewCmdShell() *cobra.Command {
	var flags struct {
		Command string
	}

	cmd := &cobra.Command{
		Use:     "shell",
		Short:   "Open a shell",
		GroupID: CommandGroupCore,
		Hidden:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return utils.RunCommand(flags.Command, "")
		},
	}

	cmd.Flags().StringVarP(&flags.Command, "command", "c", "", "Command to run")
	cmd.MarkFlagRequired("command")

	return cmd
}
