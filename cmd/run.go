package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <script>",
		Short: "Run a script",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")

			generator := func(s string) ([]byte, error) {
				name, args := utils.SplitCommand(args)
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

			runner := tui.NewCommandRunner(generator)
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
