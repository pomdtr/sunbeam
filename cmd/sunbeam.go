package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/tui"
)

var SunbeamFlags struct {
	MaxWidth  int
	MaxHeight int
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sunbeam",
	Short: "Command Line Launcher",
	Run:   Sunbeam,
}

func init() {
	rootCmd.Flags().IntVarP(&SunbeamFlags.MaxWidth, "max-width", "W", 100, "width of the window")
	rootCmd.Flags().IntVarP(&SunbeamFlags.MaxHeight, "max-height", "H", 30, "height of the window")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func Sunbeam(cmd *cobra.Command, args []string) {
	err := tui.Start(SunbeamFlags.MaxWidth, SunbeamFlags.MaxHeight)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
