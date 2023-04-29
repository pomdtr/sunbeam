package cmd

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/spf13/cobra"
)

func NewReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "read",
		Short:   "Read a file and push it",
		GroupID: coreGroupID,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return Draw(internal.NewFileGenerator(args[0]))
			}

			if !isatty.IsTerminal(os.Stdin.Fd()) {
				return Draw(internal.NewStaticGenerator(os.Stdin))
			}

			return fmt.Errorf("no input provided")
		},
	}
}
