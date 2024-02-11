package cli

import (
	"fmt"

	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/spf13/cobra"
)

func NewCmdUpgrade(cfg config.Config) *cobra.Command {
	flags := struct {
		All bool
	}{}

	cmd := &cobra.Command{
		Use:       "upgrade",
		Short:     "Upgrade sunbeam extensions",
		ValidArgs: cfg.Aliases(),
		Args:      cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !flags.All {
				return fmt.Errorf("either provide an extension or use --all")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				extensionConfig, ok := cfg.Extensions[args[0]]
				if !ok {
					return fmt.Errorf("extension %s not found", args[0])
				}

				if err := extensions.Upgrade(extensionConfig.Origin); err != nil {
					return fmt.Errorf("failed to upgrade extension: %w", err)
				}

				cmd.Printf("✅ Upgraded %s\n", args[0])
				return nil
			}

			cmd.Printf("Upgrading %d extensions...\n\n", len(cfg.Extensions))
			for alias, extensionConfig := range cfg.Extensions {
				if err := extensions.Upgrade(extensionConfig.Origin); err != nil {
					return fmt.Errorf("failed to upgrade extension %s: %w", alias, err)
				}

				cmd.Printf("✅ Upgraded %s\n", alias)
			}

			cmd.Printf("\n✅ Upgraded all extensions\n")
			return nil
		},
	}

	cmd.Flags().BoolVar(&flags.All, "all", false, "upgrade all extensions")
	return cmd
}
