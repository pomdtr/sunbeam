package cmd

import (
	"github.com/spf13/cobra"

	"github.com/SunbeamLauncher/sunbeam/app"
	"github.com/SunbeamLauncher/sunbeam/tui"
)

func Execute(version string) (err error) {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:     "sunbeam",
		Short:   "Command Line Launcher",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rootItems := make([]app.RootItem, 0)
			for _, extension := range app.Sunbeam.Extensions {
				rootItems = append(rootItems, extension.RootItems...)
			}

			for _, rootItem := range tui.Config.RootItems {
				rootItem.Subtitle = "User"
				rootItems = append(rootItems, rootItem)
			}

			rootList := tui.NewRootList(rootItems...)
			model := tui.NewModel(rootList)
			return tui.Draw(model)
		},
	}

	// Core Commands
	rootCmd.AddCommand(NewCmdExtension())
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewCmdDetail())
	rootCmd.AddCommand(NewCmdFilter())
	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(NewCmdExec())
	rootCmd.AddCommand(NewCmdServe())
	rootCmd.AddCommand(NewCmdCopy())
	rootCmd.AddCommand(NewCmdOpen())
	rootCmd.AddCommand(NewCmdBrowse())
	rootCmd.AddCommand(NewCmdSQL())

	return rootCmd.Execute()
}
