package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

func NewCmdClipboard() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clipboard",
		Short:   "Clipboard utilities",
		GroupID: "core",
	}

	cmd.AddCommand(NewCmdCopy())
	cmd.AddCommand(NewCmdPaste())

	return cmd
}

func NewCmdCopy() *cobra.Command {
	return &cobra.Command{
		Use:   "copy",
		Short: "Pipe stdin to clipboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			return clipboard.WriteAll(string(content))
		},
	}
}
func NewCmdPaste() *cobra.Command {
	return &cobra.Command{
		Use:   "paste",
		Short: "Paste clipboard content to stdout",
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
