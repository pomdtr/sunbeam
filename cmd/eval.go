package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func NewCmdEval() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "eval [file]",
		Args:   cobra.MaximumNArgs(1),
		Short:  "Evaluate a file or stdin as a page",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			generator := func() ([]byte, error) {
				buffer := bytes.Buffer{}
				i := interp.New(interp.Options{
					Stdout: &buffer,
				})
				i.Use(stdlib.Symbols)

				if len(args) > 0 {
					if _, err := i.EvalPath(args[0]); err != nil {
						return nil, err
					}
				} else {
					b, err := io.ReadAll(os.Stdin)
					if err != nil {
						return nil, err
					}

					if _, err := i.Eval(string(b)); err != nil {
						return nil, err
					}
				}

				var page types.Page
				if err := json.Unmarshal(buffer.Bytes(), &page); err != nil {
					return nil, err
				}

				return json.Marshal(page)
			}

			return Draw(generator)
		},
	}

	return cmd
}
