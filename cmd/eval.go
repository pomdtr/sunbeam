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

func parseArg(input string) any {
	if input == "" {
		return nil
	}

	var parsed any
	if err := json.Unmarshal([]byte(input), &parsed); err == nil {
		return parsed
	}

	return input
}

func NewEvalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "eval <code>",
		Short:   "Evaluate code with val.town",
		GroupID: coreGroupID,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no code provided")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			expression := types.Expression{}
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				bs, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				expression.Code = string(bs)

				expression.Args = make([]any, 0)
				for _, arg := range args {
					expression.Args = append(expression.Args, parseArg(arg))
				}
			} else {
				expression.Code = args[0]

				expression.Args = make([]any, 0)
				for _, arg := range args[1:] {
					expression.Args = append(expression.Args, parseArg(arg))
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

	cmd.Flags().StringP("expression", "e", "", "expression to evaluate")

	return cmd
}
