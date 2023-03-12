package cmd

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read [page]",
		Short: "Read page from file or stdin, and push it's content",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")

			var runner *tui.CommandRunner
			if len(args) == 0 {
				if isatty.IsTerminal(os.Stdin.Fd()) {
					exitWithErrorMsg("No input provided")
				}

				bytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Println("An error occured while reading script:", err)
					os.Exit(1)
				}

				cwd, err := os.Getwd()
				if err != nil {
					exitWithErrorMsg("could not get current working directory: %s", err)
				}

				runner = tui.NewRunner(tui.NewStaticGenerator(bytes), cwd)

			} else {

				runner = tui.NewRunner(tui.NewFileGenerator(
					args[0],
				), path.Dir(args[0]))

			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := runner.Generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Println(string(output))
				return
			}

			model := tui.NewModel(runner, tui.SunbeamOptions{
				Padding:   padding,
				MaxHeight: maxHeight,
			})

			model.Draw()
		},
	}

	return cmd
}
