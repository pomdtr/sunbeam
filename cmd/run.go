package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewRunCmd(validator tui.PageValidator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <script>",
		Short: "Run a script and push it's output",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			var extraArgs []string
			if len(args) > 1 {
				extraArgs = args[1:]
			}

			if name == "." {
				name = fmt.Sprintf("./%s", extensionBinaryName)
			}

			generator := func(s string) ([]byte, error) {
				command := exec.Command(name, extraArgs...)
				output, err := command.Output()
				if err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						return nil, fmt.Errorf("script exited with code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
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
			runner := tui.NewRunner(tui.NewCommandGenerator(name, args, cwd), validator, &url.URL{
				Scheme: "file",
				Path:   cwd,
			})
			tui.NewPaginator(runner).Draw()
		},
	}

	return cmd
}
