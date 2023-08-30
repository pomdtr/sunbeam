package cmd

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/spf13/cobra"
)

func NewReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "read",
		GroupID: coreGroupID,
		Short:   "read a sunbeam page from stdin",
		RunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				cmd.PrintErrln("no input provided")
				return cmd.Usage()
			}

			return Run(internal.NewStaticGenerator(os.Stdin))
		},
	}
}
