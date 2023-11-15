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
		Short:     "Upgrade extensions",
		GroupID:   CommandGroupCore,
		Args:      cobra.MatchAll(cobra.OnlyValidArgs, cobra.MaximumNArgs(1)),
		ValidArgs: cfg.Aliases(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if _, ok := cfg.Extensions[args[0]]; !ok {
					return fmt.Errorf("unknown extension: %s", args[0])
				}
			}

			for alias, extensionConfig := range cfg.Extensions {
				if len(args) > 0 && alias != args[0] {
					continue
				}

				if _, err := extensions.UpgradeExtension(extensionConfig); err != nil {
					return err
				}

				cmd.Printf("✅ Upgraded %s\n", alias)
			}

			return nil
		},
	}

	return cmd
}
