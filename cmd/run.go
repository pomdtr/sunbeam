package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <script>",
		Short: "Run a script and push it's output.",
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

			if check, _ := cmd.Flags().GetBool("check"); check {
				page, err := generator("")
				if err != nil {
					fmt.Println("An error occured while running the script:", err)
					os.Exit(1)
				}

				if err := schemas.Validate(page); err != nil {
					fmt.Println("Script Output is not valid:", err)
					os.Exit(1)
				}

				fmt.Println("Script Output is valid!")
				return
			}

			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "could not get current working directory:", err)
				os.Exit(1)
			}
			runner := tui.NewCommandRunner(generator, cwd)
			model := tui.NewModel(runner, tui.SunbeamOptions{
				Padding:   padding,
				MaxHeight: maxHeight,
			})
			model.Draw()
		},
	}

	cmd.Flags().Bool("check", false, "Check the script output format")

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
