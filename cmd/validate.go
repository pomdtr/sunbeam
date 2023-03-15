package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/pomdtr/sunbeam/tui"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func NewValidateCmd(validator tui.PageValidator) *cobra.Command {
	return &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate a page against the schema",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var page []byte

			if len(args) == 1 {
				if page, err = os.ReadFile(args[0]); err != nil {
					exitWithErrorMsg("Unable to read file: %s", err)
				}
			} else if !isatty.IsTerminal(os.Stdin.Fd()) {
				if page, err = io.ReadAll(os.Stdin); err != nil {
					exitWithErrorMsg("Unable to read stdin: %s", err)
				}
			} else {
				exitWithErrorMsg("No input provided")
			}

			if err := validator(page); err != nil {
				fmt.Printf("☠️ Input is not valid: %s!", err)
				os.Exit(1)
			}

			fmt.Println("✅ Input is valid!")
		},
	}
}
