package cmd

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewCmdFetch() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "fetch",
		Short:  "Fetch an extension",
		Hidden: true,
		Args:   cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				req, err := http.NewRequest(http.MethodGet, args[0], nil)
				if err != nil {
					return err
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
					return err
				}
				return nil
			}

			req, err := http.NewRequest(http.MethodPost, args[0], strings.NewReader(args[1]))
			if err != nil {
				return err
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
