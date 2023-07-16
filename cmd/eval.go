package cmd

import (
	"context"
	"encoding/json"
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
			var expression types.Expression
			if len(args) > 0 {
				expression.Code = args[0]
			} else if !isatty.IsTerminal(os.Stdin.Fd()) {
				bs, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				expression.Code = string(bs)
			} else {
				return fmt.Errorf("expression required")
			}

			if cmd.Flags().Changed("args") {
				args, _ := cmd.Flags().GetString("args")
				if err := json.Unmarshal([]byte(args), &expression.Args); err != nil {
					return err
				}
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

	cmd.Flags().StringP("args", "a", "", "Arguments to pass to the expression (JSON)")

	return cmd
}
