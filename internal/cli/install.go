package cli

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/spf13/cobra"
)

func normalizeOrigin(origin string) (string, error) {
	if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
		if _, err := os.Stat(origin); err != nil {
			return "", fmt.Errorf("failed to find origin: %w", err)
		}

		if strings.HasPrefix(origin, "~/") {
			return origin, nil
		}

		abs, err := filepath.Abs(origin)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}

		return abs, nil
	}

	return origin, nil
}

func extractAlias(origin string) (string, error) {
	originUrl, err := url.Parse(origin)
	if err != nil {
		return "", fmt.Errorf("failed to parse origin: %w", err)
	}

	base := filepath.Base(originUrl.Path)

	return strings.TrimSuffix(base, filepath.Ext(base)), nil
}

func NewCmdInstall(cfg config.Config) *cobra.Command {
	var flags struct {
		Alias string
	}

	cmd := &cobra.Command{
		Use:     "install <origin>",
		Short:   "Install sunbeam extensions",
		Aliases: []string{"add"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			origin, err := normalizeOrigin(args[0])
			if err != nil {
				return fmt.Errorf("failed to normalize origin: %w", err)
			}

			var alias string
			if flags.Alias != "" {
				alias = flags.Alias
			} else {
				a, err := extractAlias(origin)
				if err != nil {
					return fmt.Errorf("failed to get alias: %w", err)
				}
				alias = a
			}

			if _, err := extensions.LoadExtension(origin); err != nil {
				return fmt.Errorf("failed to load extension: %w", err)
			}

			if _, ok := cfg.Extensions[alias]; ok {
				return fmt.Errorf("extension %s already exists", alias)
			}

			cfg.Extensions[alias] = config.ExtensionConfig{
				Origin: origin,
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			cmd.Printf("✅ Installed %s\n", alias)
			return nil
		},
	}

	cmd.Flags().StringVar(&flags.Alias, "alias", "", "alias for extension")

	return cmd

}
