package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/spf13/cobra"
)

func NewFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch <url>",
		Short: "fetch a page and push it's output",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			method, _ := cmd.Flags().GetString("method")
			headers, _ := cmd.Flags().GetStringArray("header")
			body, _ := cmd.Flags().GetString("data")

			headerMap := make(map[string]string)
			for _, header := range headers {
				tokens := strings.SplitN(header, ":", 2)
				if len(tokens) != 2 {
					return fmt.Errorf("invalid header: %s", header)
				}

				headerMap[tokens[0]] = tokens[1]
			}

			generator := internal.NewHttpGenerator(args[0], method, headerMap, body)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					return fmt.Errorf("could not generate page: %s", err)
				}

				fmt.Println(string(output))
				return nil
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()
			return nil
		},
	}

	cmd.Flags().StringP("method", "X", "GET", "HTTP method")
	cmd.Flags().StringArrayP("header", "H", []string{}, "HTTP header")
	cmd.Flags().StringP("data", "d", "", "HTTP data")

	return cmd
}
