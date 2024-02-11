package cli

import (
	"fmt"

	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
)

func NewCmdRemove(cfg config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <alias>",
		Short:   "Remove sunbeam extensions",
		Aliases: []string{"rm", "uninstall"},
		Args:    cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			completions := cfg.Aliases()

			for _, arg := range args {
				for i, completion := range completions {
					if completion == arg {
						completions = append(completions[:i], completions[i+1:]...)
						break
					}
				}
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if _, ok := cfg.Extensions[arg]; !ok {
					return fmt.Errorf("extension %s not found", arg)
				}

				delete(cfg.Extensions, arg)
				var root []sunbeam.Action
				for _, action := range cfg.Root {
					if action.Extension == arg {
						continue
					}

					root = append(root, action)
				}

				cfg.Root = root
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			if len(args) == 1 {
				cmd.Printf("✅ Removed %s\n", args[0])
				return nil
			}

			cmd.Printf("✅ Removed %d extensions\n", len(args))
			return nil
		},
	}
}
