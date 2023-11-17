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
			if len(args) > 0 {
				extension, ok := cfg.Extensions[args[0]]
				if !ok {
					return fmt.Errorf("extension %s not found", args[0])
				}

				if err := extensions.Upgrade(extension); err != nil {
					return fmt.Errorf("failed to upgrade extension: %w", err)
				}

				cmd.Printf("✅ Upgraded %s\n", args[0])
				return nil
			}

			cmd.Printf("Upgrading %d extensions...\n\n", len(cfg.Extensions))
			for alias, extension := range cfg.Extensions {
				if err := extensions.Upgrade(extension); err != nil {
					return fmt.Errorf("failed to upgrade extension %s: %w", alias, err)
				}

				cmd.Printf("✅ Upgraded %s\n", alias)
			}

			cmd.Printf("\n✅ Upgraded all extensions\n")
			return nil
		},
	}

	return cmd
}
