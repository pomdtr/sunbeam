package cmd

import (
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func NewCmdPaste() *cobra.Command {
	return &cobra.Command{
		Use:     "paste",
		Short:   "Paste text from clipboard to stdout",
		GroupID: CommandGroupCore,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := clipboard.ReadAll()
			if err != nil {
				return err
			}

			if _, err := os.Stdout.WriteString(output); err != nil {
				return err
			}

			return nil
		},
	}

}
