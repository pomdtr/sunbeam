package cmd

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func NewPasteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "paste",
		Short: "Paste system clipboard to stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			text, err := clipboard.ReadAll()
			if err != nil {
				return fmt.Errorf("unable to read from clipboard: %s", err)
			}

			fmt.Print(text)
			return nil
		},
	}
}
