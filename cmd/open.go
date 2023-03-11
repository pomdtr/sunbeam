package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func NewOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <url>",
		Short: "Open file or url in default application",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := browser.OpenURL(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Unable to open link:", err)
				os.Exit(1)
			}
		},
	}
}
