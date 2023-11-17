package cmd

import (
	"fmt"

	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/spf13/cobra"
)

func NewCmdUpgrade(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "upgrade",
		Short:     "Upgrade sunbeam",
		GroupID:   CommandGroupCore,
		ValidArgs: cfg.Aliases(),
		Args:      cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, alias := range cfg.Aliases() {
				if len(args) == 1 && alias != args[0] {
					continue
				}

				extensionConfig, ok := cfg.Extensions[alias]
				if !ok {
					return fmt.Errorf("extension %s not found", alias)
				}

				cmd.Printf("Upgrading %s\n", alias)
				if err := extensions.Upgrade(extensionConfig); err != nil {
					cmd.PrintErrf("Failed to upgrade %s: %s\n", alias, err)
					return err
				}

				fmt.Printf("Upgraded %s\n", alias)
			}

			return nil
		},
	}

	return cmd
}
