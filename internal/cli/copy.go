package cli

import (
	"io"
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func NewCmdCopy() *cobra.Command {
	return &cobra.Command{
		Use:     "copy",
		Short:   "Copy text from stdin or paste text to stdout",
		GroupID: CommandGroupCore,
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			return clipboard.WriteAll(string(input))
		},
	}

}
