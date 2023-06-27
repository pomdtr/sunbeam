package cmd

import (
	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewEvalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eval <code>",
		Short: "Evaluate code with val.town",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			expression := types.Expression{
				Code: args[0],
			}
			return Run(internal.NewRequestGenerator(expression.Request()))
		},
	}

	return cmd
}
