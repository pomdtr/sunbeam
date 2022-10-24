package cmd

import (
	"log"
	"strings"

	"github.com/pomdtr/sunbeam/api"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

type RunFlags struct {
	Params []string
}

var runFlags = RunFlags{}

func init() {
	commands := api.Commands
	validArgs := make([]string, len(commands))
	for i, command := range commands {
		validArgs[i] = command.Target()
	}

	runCmd.ValidArgs = validArgs
	runCmd.Flags().StringArrayVarP(&runFlags.Params, "param", "p", []string{}, "Parameters to pass to the script")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a sunbeam script",
	Args:  cobra.ExactArgs(1),
	Run:   sunbeamRun,
}

func sunbeamRun(cmd *cobra.Command, args []string) {
	params := make(map[string]string)
	for _, param := range runFlags.Params {
		tokens := strings.SplitN(param, "=", 2)
		params[tokens[0]] = tokens[1]
	}

	err := tui.Run(args[0], params)
	if err != nil {
		log.Fatalf("could not run script: %v", err)
	}
}
