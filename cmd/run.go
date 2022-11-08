package cmd

import (
	"fmt"

	"github.com/pomdtr/sunbeam/api"
	"github.com/spf13/cobra"
)

type RunFlags struct {
	Params []string
}

var runFlags = RunFlags{}

func init() {
	validargs := make([]string, 0)
	for _, manifest := range api.Sunbeam.Extensions {
		for scriptName := range manifest.Scripts {
			validargs = append(validargs, fmt.Sprintf("%s.%s", manifest.Name, scriptName))
		}
	}

	runCmd.ValidArgs = validargs
	runCmd.Flags().StringArrayVarP(&runFlags.Params, "param", "p", []string{}, "Parameters to pass to the script")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run-script",
	Short: "Run a sunbeam script",
	Args:  cobra.ExactArgs(1),
	// Run:   sunbeamRun,
}

// func sunbeamRun(cmd *cobra.Command, args []string) {
// 	params := make(map[string]string)
// 	for _, param := range runFlags.Params {
// 		tokens := strings.SplitN(param, "=", 2)
// 		params[tokens[0]] = tokens[1]
// 	}

// 	tokens := strings.SplitN(args[0], ".", 2)
// 	extensionName, scriptName := tokens[0], tokens[1]
// 	command, ok := api.Sunbeam.GetScript(extensionName, scriptName)
// 	if !ok {
// 		log.Fatalf("Command %s.%s not found", extensionName, scriptName)
// 	}

// 	err := tui.Run(command, params)
// 	if err != nil {
// 		log.Fatalf("could not run script: %v", err)
// 	}
// }
