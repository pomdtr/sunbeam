package cmd

import (
	"fmt"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func NewOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <url>",
		Short: "Open file or url in default application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := browser.OpenURL(args[0])
			if err != nil {
				return fmt.Errorf("unable to open link: %s", err)
			}
			return nil
		},
	}
}
