package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/schemas"
	"github.com/spf13/cobra"
)

func NewValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate a page against the schema",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input []byte

			if len(args) == 1 {
				c, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("unable to read file: %s", err)
				}
				input = c
			} else if !isatty.IsTerminal(os.Stdin.Fd()) {
				c, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("unable to read stdin: %s", err)
				}

				input = c
			} else {
				return fmt.Errorf("no input provided")
			}

			var v any
			if err := json.Unmarshal(input, &v); err != nil {
				return fmt.Errorf("unable to parse input: %s", err)
			}

			if err := schemas.Validate(v); err != nil {
				return fmt.Errorf("input is not valid: %s", err)
			}

			fmt.Println("✅ Input is valid!")
			return nil
		},
	}
}
