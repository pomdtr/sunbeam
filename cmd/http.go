package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func NewHttpCmd() *cobra.Command {
	flags := struct {
		Method string
		Data   string
		Header []string
	}{}

	cmd := &cobra.Command{
		Use:   "http",
		Short: "Simple http client",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			target, err := url.Parse(args[0])
			if err != nil {
				exitWithErrorMsg("could not parse url: %s", err)
			}

			if target.Scheme == "" {
				exitWithErrorMsg("missing scheme: %s", args[0])
			}

			if target.Scheme != "http" && target.Scheme != "https" {
				exitWithErrorMsg("invalid scheme: %s", target.Scheme)
			}

			var body io.Reader
			if flags.Data != "" {
				if flags.Method == "" {
					flags.Method = "POST"
				}

				if flags.Method != "POST" && flags.Method != "PUT" {
					exitWithErrorMsg("invalid method: %s", flags.Method)
				}

				body, err = extractBody(flags.Data)
				if err != nil {
					exitWithErrorMsg("could not extract body: %s", err)
				}
			}

			if flags.Method == "" {
				flags.Method = "GET"
			}

			req, err := http.NewRequest(flags.Method, args[0], body)
			if err != nil {
				exitWithErrorMsg("could not create request: %s", err)
			}

			for _, header := range flags.Header {
				tokens := strings.SplitN(header, ":", 2)
				if len(tokens) != 2 {
					exitWithErrorMsg("invalid header: %s", header)
				}
				req.Header.Set(tokens[0], tokens[1])
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				exitWithErrorMsg("could not send request: %s", err)
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				exitWithErrorMsg("request failed: %s", resp.Status)
			}

			if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
				exitWithErrorMsg("could not read response: %s", err)
			}

		},
	}

	cmd.Flags().StringVarP(&flags.Method, "method", "X", "", "HTTP method")
	cmd.Flags().StringVarP(&flags.Data, "data", "d", "", "HTTP request body")
	cmd.Flags().StringArrayVarP(&flags.Header, "header", "H", nil, "HTTP request header")

	return cmd
}

func extractBody(data string) (io.Reader, error) {
	if !strings.HasPrefix(data, "@") {
		return strings.NewReader(data), nil
	}

	filepath := strings.TrimPrefix(data, "@")
	if filepath == "-" {
		if isatty.IsTerminal(os.Stdin.Fd()) {
			return nil, fmt.Errorf("cannot read from stdin")
		}

		return os.Stdin, nil
	}

	return os.Open(filepath)
}
