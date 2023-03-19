package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewHTTPCmd(validator tui.PageValidator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http <url>",
		Short: "Run a HTTP server",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			method, _ := cmd.Flags().GetString("method")
			headers, _ := cmd.Flags().GetStringArray("header")
			body, _ := cmd.Flags().GetString("data")

			target, err := url.Parse(args[0])
			if err != nil {
				exitWithErrorMsg("invalid url: %s", err)
			}

			headerMap := make(map[string]string)
			for _, header := range headers {
				tokens := strings.SplitN(header, ":", 2)
				if len(tokens) != 2 {
					exitWithErrorMsg("invalid header: %s", header)
				}

				headerMap[tokens[0]] = tokens[1]
			}

			generator := tui.NewHttpGenerator(args[0], method, headerMap, body)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Println(string(output))
				return
			}

			runner := tui.NewRunner(generator, validator, &url.URL{
				Scheme: target.Scheme,
				Host:   target.Host,
				Path:   path.Dir(target.Path),
			})
			tui.NewPaginator(runner).Draw()
		},
	}

	cmd.Flags().StringP("method", "X", "GET", "HTTP method")
	cmd.Flags().StringArrayP("header", "H", []string{}, "HTTP header")
	cmd.Flags().StringP("data", "d", "", "HTTP data")

	return cmd
}
