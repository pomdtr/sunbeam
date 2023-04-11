package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pomdtr/sunbeam/internal"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func NewReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <page>",
		Short: "Read page from file or stdin, and push it's content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			generator := internal.NewFileGenerator(args[0])
			if !isatty.IsTerminal(os.Stderr.Fd()) {
				output, err := generator()
				if err != nil {
					return fmt.Errorf("could not generate page: %s", err)
				}

				json.NewDecoder(os.Stderr).Decode(output)
				return nil
			}

			runner := internal.NewRunner(generator)
			model := internal.NewPaginator(runner)

			model.Draw()
			return nil
		},
	}

	return cmd
}
