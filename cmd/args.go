package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed minimist.sh
var minimist string

func NewArgsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "args",
		Short: "Parse command line arguments",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(minimist)
			return nil
		},
	}
}
