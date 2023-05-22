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
		Use:     "read [file]",
		Short:   "Read a a page from a file and push it",
		GroupID: coreGroupID,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return Run(internal.NewFileGenerator(args[0]))
			}

			if !isatty.IsTerminal(os.Stdin.Fd()) {
				return Run(internal.NewStaticGenerator(os.Stdin))
			}

			return fmt.Errorf("no input provided")
		},
	}
}
