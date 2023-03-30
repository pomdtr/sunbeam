package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <script>",
		Short: "Run a script and push it's output",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cwd, _ := os.Getwd()

			command := strings.Join(args, " ")
			generator := tui.NewCommandGenerator(command, "", cwd)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			runner := tui.NewRunner(generator)
			tui.NewPaginator(runner).Draw()
		},
	}

	return cmd
}
