package cmd

import (
	"io"
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func NewCmdCopy() *cobra.Command {
	return &cobra.Command{
		Use:   "copy",
		Short: "Copy text from stdin",
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			return clipboard.WriteAll(string(content))
		},
	}
}
