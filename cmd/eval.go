package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewEvalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eval <code>",
		Short: "Evaluate code with val.town",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no code provided")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var code string
			if len(args) > 0 {
				code = args[0]
			} else {
				bs, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				code = string(bs)
			}

			expression := types.Expression{
				Code: code,
			}
			return Run(internal.NewRequestGenerator(expression.Request()))
		},
	}

	return cmd
}
