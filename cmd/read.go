package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read [page]",
		Short: "Read page from file or stdin, and push it's content",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				if isatty.IsTerminal(os.Stdin.Fd()) {
					return fmt.Errorf("no input provided")
				}

				var content []byte
				format, _ := cmd.Flags().GetString("format")
				runner := internal.NewRunner(func() ([]byte, error) {
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
						var page types.Page
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
				})

				internal.NewPaginator(runner).Draw()
				return nil

			}

			var generator internal.PageGenerator
			// If the format flag is set, we override the detected format
			if cmd.Flags().Changed("format") {
				format, _ := cmd.Flags().GetString("format")
				if format == "yaml" || format == "yml" {
					generator = func() ([]byte, error) {
						bytes, err := os.ReadFile(args[0])
						if err != nil {
							return nil, err
						}

						var page types.Page
						if err := yaml.Unmarshal(bytes, &page); err != nil {
							return nil, err
						}

						return json.Marshal(page)
					}
				} else {
					generator = func() ([]byte, error) {
						return os.ReadFile(args[0])
					}
				}
			} else {
				// By default, we detect the format of the file based on the extension
				generator = internal.NewFileGenerator(args[0])
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator()
				if err != nil {
					return fmt.Errorf("could not generate page: %s", err)
				}

				fmt.Println(string(output))
				return nil
			}

			runner := internal.NewRunner(generator)
			model := internal.NewPaginator(runner)

			model.Draw()
			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "json", "Format of the input file. Can be json or yaml.")
	return cmd
}
