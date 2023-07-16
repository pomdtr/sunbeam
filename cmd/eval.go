package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewEvalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "eval [expression]",
		Short:   "Evaluate code with val.town",
		GroupID: coreGroupID,
		RunE: func(cmd *cobra.Command, args []string) error {
			var code string
			if len(args) > 0 {
				code = args[0]
			} else if !isatty.IsTerminal(os.Stdin.Fd()) {
				bs, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				code = string(bs)
			} else {
				return fmt.Errorf("expression required")
			}

			expression := types.Expression{
				Code: code,
			}

			bs, err := expression.Request().Do(context.Background())
			if err != nil {
				return err
			}

			if _, err := os.Stdout.Write(bs); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
