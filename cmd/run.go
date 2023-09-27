package cmd

import (
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/spf13/cobra"
)

func NewCmdRun(config tui.Config) *cobra.Command {
	extensions := tui.NewExtensions(config.Aliases)

	cmd := &cobra.Command{
		Use:                "run <origin> [args...]",
		Short:              "Run an extension without installing it",
		Args:               cobra.MinimumNArgs(1),
		GroupID:            coreGroupID,
		DisableFlagParsing: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			manifest, err := extensions.Get(args[0])
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			if len(args) == 1 {
				commands := make([]string, 0)
				for _, command := range manifest.Commands {
					if command.Hidden {
						continue
					}
					commands = append(commands, command.Name)
				}

				return commands, cobra.ShellCompDirectiveNoFileComp
			}

			return []string{}, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			rootCmd, err := NewCustomCommand(args[0], config)
			if err != nil {
				return err
			}

			rootCmd.Use = args[0]
			rootCmd.SetArgs(args[1:])
			return rootCmd.Execute()
		},
	}

	return cmd
}
