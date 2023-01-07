package cmd

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func NewCmdPaste() *cobra.Command {
	return &cobra.Command{
		Use:     "paste",
		Short:   "Paste file to clipboard",
		GroupID: "core",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clipboadContent, err := clipboard.ReadAll()
			if err != nil {
				return fmt.Errorf("failed to read clipboard: %w", err)
			}

			os.Stdout.WriteString(clipboadContent)
			return nil
		},
	}
}
