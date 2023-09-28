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
		Use:   "fetch",
		Short: "Fetch an extension",
		Args:  cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				resp, err := http.Get(args[0])
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
					return err
				}
				return nil
			}

			resp, err := http.Post(args[0], "application/json", strings.NewReader(args[1]))
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
