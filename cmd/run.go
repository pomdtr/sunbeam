package cmd

import (
	"log"

	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a sunbeam script",
	Args:  cobra.ExactArgs(1),
	Run:   sunbeamRun,
}

func sunbeamRun(cmd *cobra.Command, args []string) {
	err := tui.Run(args[0])
	if err != nil {
		log.Fatalf("could not run script: %v", err)
	}

}
