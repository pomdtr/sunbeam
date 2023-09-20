package cmd

import (
	"os"

	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/spf13/cobra"
)

func NewCmdUpdate(config tui.Config, cachePath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update extension cache",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = os.Remove(cachePath)
			extensions, err := tui.LoadExtensions(config, cachePath)
			if err != nil {
				return err
			}

			cmd.PrintErrf("Refreshed %d extensions", len(extensions))
			return nil
		},
	}

	return cmd
}
