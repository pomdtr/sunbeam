package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewCmdFetch() *cobra.Command {
	flags := struct {
		headers []string
		method  string
		data    string
		user    string
	}{}
	cmd := &cobra.Command{
		Use:     "fetch <url> [body]",
		Short:   "Simple http client inspired by curl",
		GroupID: CommandGroupDev,
		Args:    cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			origin, err := url.Parse(args[0])
			if err != nil {
				return err
			}

			var method string
			if flags.method != "" {
				method = flags.method
			} else if flags.data != "" {
				method = http.MethodPost
			} else {
				method = http.MethodGet
			}

			var body io.Reader
			if flags.data != "" {
				if method == http.MethodGet {
					return fmt.Errorf("cannot send body with GET request")
				}

				if flags.data == "@-" {
					body = os.Stdin
				} else if strings.HasPrefix(flags.data, "@") {
					file, err := os.Open(flags.data[1:])
					if err != nil {
						return fmt.Errorf("failed to open file: %w", err)
					}

					body = file
				} else {
					body = strings.NewReader(flags.data)
				}
			}

			req, err := http.NewRequest(method, origin.String(), body)
			if err != nil {
				return err
			}

			for _, v := range flags.headers {
				parts := strings.SplitN(v, ":", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid header format, expected key:value")
				}

				req.Header.Add(parts[0], parts[1])
			}

			if flags.user != "" {
				parts := strings.SplitN(flags.user, ":", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid user format, expected user:pass")
				}

				req.SetBasicAuth(parts[0], parts[1])
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
				return err
			}

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to fetch url: %s", resp.Status)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&flags.headers, "header", "H", nil, "HTTP headers to add to the request")
	cmd.Flags().StringVarP(&flags.method, "method", "X", "", "HTTP method to use")
	cmd.Flags().StringVarP(&flags.data, "data", "d", "", "HTTP body to send. Use @- to read from stdin, or @<file> to read from a file.")
	cmd.Flags().StringVarP(&flags.user, "user", "u", "", "HTTP basic auth to use")

	return cmd
}
