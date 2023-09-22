package cmd

import (
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/spf13/cobra"
)

func NewCmdRun(config tui.Config) *cobra.Command {
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

			extension, err := tui.LoadExtension(args[0])
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			if len(args) == 1 {
				commands := make([]string, 0)
				for _, command := range extension.Manifest.Commands {
					commands = append(commands, command.Name)
				}

				return commands, cobra.ShellCompDirectiveNoFileComp
			}

			return []string{}, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "--help" || args[0] == "-h" {
				return cmd.Help()
			}

			if len(args) == 1 {
				return tui.Draw(tui.NewExtensionPage(args[0]), config.Window)
			}

			extension, err := tui.LoadExtension(args[0])
			if err != nil {
				return err
			}

			scriptCmd, err := NewCustomCmd(extension, config)
			if err != nil {
				return err
			}
			scriptCmd.SilenceErrors = true
			scriptCmd.SilenceUsage = true

			scriptCmd.SetArgs(args[1:])
			return scriptCmd.Execute()
		},
	}

	return cmd
}
