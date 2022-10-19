package cmd

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a sunbeam script",
	Args:  cobra.MaximumNArgs(1),
	Run:   sunbeamRun,
}

func sunbeamRun(cmd *cobra.Command, args []string) {
}
