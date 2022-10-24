package cmd

import (
	"log"

	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func init() {
	commands := api.Commands
	validArgs := make([]string, len(commands))
	for i, command := range commands {
		validArgs[i] = command.Target()
	}

	runCmd.ValidArgs = validArgs
	rootCmd.AddCommand(runCmd)
}

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
