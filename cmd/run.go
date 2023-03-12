package cmd

import (
	"fmt"
	"os"
	"os/exec"

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
			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")

			generator := func(s string) ([]byte, error) {
				name, args := SplitCommand(args)
				command := exec.Command(name, args...)
				output, err := command.Output()
				if err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						return nil, fmt.Errorf("Script exited with code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
					}

					return nil, err
				}

				return output, nil
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			cwd, _ := os.Getwd()
			command, args := SplitCommand(args)
			runner := tui.NewRunner(tui.NewCommandGenerator(command, args, cwd), cwd)
			tui.NewModel(runner, tui.SunbeamOptions{
				Padding:   padding,
				MaxHeight: maxHeight,
			}).Draw()
		},
	}

	return cmd
}

func SplitCommand(fields []string) (string, []string) {
	if len(fields) == 0 {
		return "", nil
	}

	if len(fields) == 1 {
		return fields[0], nil
	}

	return fields[0], fields[1:]
}
