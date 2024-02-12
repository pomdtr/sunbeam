package cli

import (
	"fmt"

	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/spf13/cobra"
)

func NewCmdRun(cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <origin>",
		Short:              "Run remote extension",
		Args:               cobra.MinimumNArgs(1),
		Aliases:            []string{"add"},
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			origin, err := normalizeOrigin(args[0])
			if err != nil {
				return fmt.Errorf("failed to normalize origin: %w", err)
			}

			extension, err := extensions.LoadExtension(origin)
			if err != nil {
				return fmt.Errorf("failed to load extension: %w", err)
			}
			extension.Env = cfg.Env

			cfg.Extensions[args[0]] = config.ExtensionConfig{
				Origin: origin,
			}

			command, err := NewCmdCustom(args[0], extension, cfg)
			if err != nil {
				return fmt.Errorf("failed to create command: %w", err)
			}

			command.SetArgs(args[1:])
			return command.Execute()
		},
	}

	return cmd
}
