package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func NewCopyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "copy",
		Short: "Copy stdin to system clipboard",
		Run: func(cmd *cobra.Command, args []string) {
			text, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Unable to read stdin:", err)
				os.Exit(1)
			}

			if err := clipboard.WriteAll(string(text)); err != nil {
				fmt.Fprintln(os.Stderr, "Unabble to write to clipboard:", err)
				os.Exit(1)
			}
		},
	}
}
