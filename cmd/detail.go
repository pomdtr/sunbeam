package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewDetailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "detail",
		Short:   "parse text from stdin",
		Args:    cobra.NoArgs,
		GroupID: coreGroupID,
		RunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no input provided")
			}

			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("could not read input: %s", err)
			}
			text := string(b)

			title := "Sunbeam"
			if cmd.Flags().Changed("title-row") {
				var parts []string
				if runtime.GOOS == "windows" {
					parts = strings.SplitN(text, "\r\n", 2)
				} else {
					parts = strings.SplitN(text, "\n", 2)
				}
				if len(parts) > 0 {
					title = string(parts[0])
				}
			}
			if cmd.Flags().Changed("title") {
				title, _ = cmd.Flags().GetString("title")
			}

			return Run(func() (*types.Page, error) {
				return &types.Page{
					Title: title,
					Type:  "detail",
					Text:  text,
				}, nil
			})
		},
	}

	cmd.Flags().StringP("title", "t", "Sunbeam", "title of the page")
	cmd.Flags().Bool("title-row", false, "use first row as title")
	cmd.MarkFlagsMutuallyExclusive("title", "title-row")

	return cmd
}
