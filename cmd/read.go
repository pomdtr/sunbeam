package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read [page]",
		Short: "Read page from file or stdin, and push it's content",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var runner *tui.CommandRunner
			if len(args) == 0 {
				if isatty.IsTerminal(os.Stdin.Fd()) {
					exitWithErrorMsg("No input provided")
				}

				cwd, err := os.Getwd()
				if err != nil {
					exitWithErrorMsg("could not get current working directory: %s", err)
				}

				var content []byte
				format, _ := cmd.Flags().GetString("format")

				runner = tui.NewRunner(func(input string) ([]byte, error) {
					if content != nil {
						return content, nil
					}

					// We only read from stdin once
					// We do it from the generator function since it will run in a tea.Cmd
					bytes, err := io.ReadAll(os.Stdin)
					if err != nil {
						return nil, err
					}

					if format == "yaml" || format == "yml" {
						var page schemas.Page
						if err := yaml.Unmarshal(bytes, &page); err != nil {
							return nil, err
						}
						content, err = json.Marshal(page)
						if err != nil {
							return nil, err
						}
					} else {
						content = bytes
					}

					return content, err
				}, cwd)

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

			model := tui.NewModel(runner)

			model.Draw()
		},
	}

	cmd.Flags().StringP("format", "f", "json", "Format of the input file. Can be json or yaml.")
	return cmd
}
