package cmd

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func NewPasteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "paste",
		Short: "Paste system clipboard to stdout",
		Run: func(cmd *cobra.Command, args []string) {
			text, err := clipboard.ReadAll()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Unable to read from clipboard:", err)
				os.Exit(1)
			}

			fmt.Print(text)
		},
	}
}
