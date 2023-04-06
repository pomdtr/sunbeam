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
		RunE: func(cmd *cobra.Command, args []string) error {
			text, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("unable to read stdin: %s", err)
			}

			if err := clipboard.WriteAll(string(text)); err != nil {
				return fmt.Errorf("unable to write to clipboard: %s", err)
			}

			return nil
		},
	}
}
