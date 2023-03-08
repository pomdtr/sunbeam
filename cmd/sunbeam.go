package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/tui"
)

func Execute(version string) (err error) {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:     "sunbeam <script>",
		Short:   "Command Line Launcher",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var runner *tui.CommandRunner
			if len(args) == 0 {
				tempfile, err := os.CreateTemp("", "sunbeam_input.json")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to create tempfile: %v", err)
					os.Exit(1)
				}
				defer os.Remove(tempfile.Name())

				_, err = io.Copy(tempfile, os.Stdin)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to copy stdin to tempfile: %v", err)
					os.Exit(1)
				}

				runner = tui.NewCommandRunner("cat", tempfile.Name())
			} else {
				command := args[0]
				var commandArgs []string
				if len(args) > 1 {
					commandArgs = args[1:]
				}
				runner = tui.NewCommandRunner(command, commandArgs...)
			}

			model := tui.NewModel(runner)
			return tui.Draw(model)
		},
	}

	return rootCmd.Execute()
}
