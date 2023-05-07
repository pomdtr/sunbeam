package cmd

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func NewFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fetch",
		GroupID: coreGroupID,
		Short:   "fetch http using a curl-like syntax",
		RunE: func(cmd *cobra.Command, args []string) error {
			method, _ := cmd.Flags().GetString("method")
			headers, _ := cmd.Flags().GetStringArray("header")
			user, _ := cmd.Flags().GetString("user")

			var body io.Reader
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				body = os.Stdin
			}

			if method == "" {
				if body != nil {
					method = "POST"
				} else {
					method = "GET"
				}
			}

			req, err := http.NewRequest(method, args[0], body)
			if err != nil {
				return err
			}

			for _, header := range headers {
				parts := strings.SplitN(header, ":", 2)
				if len(parts) != 2 {
					return err
				}

				req.Header.Set(parts[0], parts[1])
			}

			if user != "" {
				parts := strings.SplitN(user, ":", 2)
				if len(parts) != 2 {
					return err
				}

				req.SetBasicAuth(parts[0], parts[1])
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			io.Copy(os.Stdout, resp.Body)

			return nil
		},
	}

	cmd.Flags().String("method", "GET", "HTTP method")
	cmd.Flags().StringArrayP("header", "H", []string{}, "HTTP header")
	cmd.Flags().String("user", "", "HTTP user")

	return cmd
}
