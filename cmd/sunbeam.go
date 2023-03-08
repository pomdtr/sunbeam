package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/scripts"
	"github.com/pomdtr/sunbeam/tui"
)

func Execute(version string) (err error) {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam <script>",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

You will need to provide a compatible script as the first argument to you use sunbeam. See http://pomdtr.github.io/sunbeam for more information.`,
		Args:    cobra.MinimumNArgs(1),
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			var runner *tui.CommandRunner
			command := args[0]
			var commandArgs []string
			if len(args) > 1 {
				commandArgs = args[1:]
			}

			if check, _ := cmd.Flags().GetBool("check"); check {
				cmd := exec.Command(command, commandArgs...)
				output, err := cmd.Output()
				if err != nil {
					fmt.Println("An error occured while running the script:", err)
					os.Exit(1)
				}

				var v interface{}
				if err := json.Unmarshal(output, &v); err != nil {
					fmt.Println("Script is not valid:", err)
					os.Exit(1)
				}

				if err := scripts.Schema.Validate(v); err != nil {
					fmt.Println("Script is not valid:", err)
					os.Exit(1)
				}

				fmt.Println("Script is valid!")
				os.Exit(0)
			}

			runner = tui.NewCommandRunner(command, commandArgs...)
			model := tui.NewModel(runner)
			tui.Draw(model)
		},
	}

	rootCmd.Flags().Bool("check", false, "Check if the script is valid")

	return rootCmd.Execute()
}
