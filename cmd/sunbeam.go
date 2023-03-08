package cmd

import (
	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/tui"
)

func Execute(version string) (err error) {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam <script>",
		Short:        "Command Line Launcher",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		Version:      version,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			command := args[0]
			var commandArgs []string
			if len(args) > 1 {
				commandArgs = args[1:]
			}

			model := tui.NewModel(tui.NewCommandRunner(command, commandArgs))
			return tui.Draw(model)
		},
	}

	return rootCmd.Execute()
}
