package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewCmdDocs() *cobra.Command {
	return &cobra.Command{
		Use:    "docs",
		Args:   cobra.ExactArgs(1),
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			if _, err := os.Stat(target); os.IsNotExist(err) {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
			root := cmd.Root()
			root.DisableAutoGenTag = true
			return doc.GenMarkdownTreeCustom(
				root,
				target,
				func(s string) string {
					return ""
				},
				func(s string) string { return fmt.Sprintf("./%s", s) },
			)
		},
	}
}
