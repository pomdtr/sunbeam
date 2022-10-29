package cmd

import (
	"fmt"
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

	validargs := make([]string, 0)
	for _, manifest := range api.Sunbeam.Extensions {
		for commandName := range manifest.Commands {
			validargs = append(validargs, fmt.Sprintf("%s.%s", manifest.Name, commandName))
		}
	}

	runCmd.ValidArgs = validargs
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

	tokens := strings.SplitN(args[0], ".", 2)
	extensionName, commandName := tokens[0], tokens[1]
	command, ok := api.Sunbeam.GetCommand(extensionName, commandName)
	if !ok {
		log.Fatalf("Command %s.%s not found", extensionName, commandName)
	}

	err := tui.Run(command, params)
	if err != nil {
		log.Fatalf("could not run script: %v", err)
	}
}
