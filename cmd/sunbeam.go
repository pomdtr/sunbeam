package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"

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

			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")

			runner = tui.NewCommandRunner(command, commandArgs...)
			model := tui.NewModel(runner, tui.SunbeamOptions{
				Padding:   padding,
				MaxHeight: maxHeight,
			})
			tui.Draw(model, maxHeight == 0)
		},
	}

	rootCmd.Flags().IntP("padding", "p", lookupInt("SUNBEAM_PADDING", 0), "padding around the window")
	rootCmd.Flags().IntP("height", "H", lookupInt("SUNBEAM_HEIGHT", 0), "maximum height of the window")
	rootCmd.Flags().Bool("check", false, "check if the script is valid")

	return rootCmd.Execute()
}

func lookupInt(key string, fallback int) int {
	if env, ok := os.LookupEnv(key); ok {
		if value, err := strconv.Atoi(env); err == nil {
			return value
		}
	}

	return fallback
}
