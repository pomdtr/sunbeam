package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/tui"
)

var SunbeamFlags struct {
	Height int
	Width  int
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sunbeam",
	Short: "Command Line Launcher",
	Run:   Sunbeam,
}

func init() {
	rootCmd.Flags().IntVarP(&SunbeamFlags.Height, "height", "H", 33, "height of the window")
	rootCmd.Flags().IntVarP(&SunbeamFlags.Width, "width", "W", 105, "width of the window")
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
	err := tui.Start(SunbeamFlags.Width, SunbeamFlags.Height)
	if err != nil {
		log.Fatalf("could not start tui: %v", err)
	}
}
