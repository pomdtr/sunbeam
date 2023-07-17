package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func NewFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fetch <url>",
		GroupID: coreGroupID,
		Args:    cobra.ExactArgs(1),
		Short:   "Fetch http using a curl-like syntax",
		RunE: func(cmd *cobra.Command, args []string) error {
			headersFlag, _ := cmd.Flags().GetStringArray("header")

			var body []byte
			if cmd.Flags().Changed("data") {
				data, _ := cmd.Flags().GetString("data")
				body = []byte(data)
			} else if !isatty.IsTerminal(os.Stdin.Fd()) {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				body = b
			}

			var method string
			if cmd.Flags().Changed("method") {
				method, _ = cmd.Flags().GetString("method")
			} else if len(body) > 0 {
				method = http.MethodPost
			} else {
				method = http.MethodGet
			}

			if method == "GET" && len(body) > 0 {
				return fmt.Errorf("cannot specify request body for GET request")
			}

			req, err := http.NewRequest(method, args[0], bytes.NewReader(body))
			if err != nil {
				return err
			}

			query := req.URL.Query()
			queryParams, _ := cmd.Flags().GetStringArray("query")
			for _, param := range queryParams {
				parts := strings.SplitN(param, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid query: %s", query)
				}

				query.Add(parts[0], parts[1])
			}

			for _, header := range headersFlag {
				parts := strings.SplitN(header, ":", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid header: %s", header)
				}

				req.Header.Set(parts[0], parts[1])
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				return fmt.Errorf("request failed: %s", resp.Status)
			}

			if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringP("method", "X", "", "http method")
	cmd.Flags().StringArrayP("header", "H", []string{}, "http header")
	cmd.Flags().StringArrayP("query", "q", []string{}, "http query param")
	cmd.Flags().StringP("data", "d", "", "http request body")

	return cmd
}
