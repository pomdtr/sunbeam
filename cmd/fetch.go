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
	cmd := &cobra.Command{
		Use:    "fetch <url> [body]",
		Short:  "Simple http client, mostly used for accessing sunbeam scripts remotely.",
		Hidden: true,
		Args:   cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			origin, err := url.Parse(args[0])
			if err != nil {
				return err
			}

			var token string
			if origin.User != nil {
				token = origin.User.String()
			}

			if len(args) == 1 {
				req, err := http.NewRequest(http.MethodGet, args[0], nil)
				if err != nil {
					return err
				}

				if token != "" {
					req.Header.Set("Authorization", "Bearer "+token)
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("failed to fetch extension: %s", resp.Status)
				}

				if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
					return err
				}
				return nil
			}

			req, err := http.NewRequest(http.MethodPost, args[0], strings.NewReader(args[1]))
			if err != nil {
				return err
			}
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to fetch extension: %s", resp.Status)
			}

			if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
