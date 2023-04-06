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
			var err error
			var page []byte

			if len(args) == 1 {
				if page, err = os.ReadFile(args[0]); err != nil {
					return fmt.Errorf("unable to read file: %s", err)
				}
			} else if !isatty.IsTerminal(os.Stdin.Fd()) {
				if page, err = io.ReadAll(os.Stdin); err != nil {
					return fmt.Errorf("unable to read stdin: %s", err)
				}
			} else {
				return fmt.Errorf("no input provided")
			}

			var v any
			if err := json.Unmarshal(page, &v); err != nil {
				return fmt.Errorf("unable to parse input: %s", err)
			}

			if err := schemas.Validate(page); err != nil {
				return fmt.Errorf("input is not valid: %s", err)
			}

			fmt.Println("✅ Input is valid!")
			return nil
		},
	}
}
