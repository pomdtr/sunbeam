package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/cli/browser"
	"github.com/spf13/cobra"
)

func NewCmdBrowse() *cobra.Command {
	return &cobra.Command{
		Use:   "browse",
		Short: "Browse url with the default browser",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var input string
			if len(args) == 0 {
				bytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				input = strings.Trim(string(bytes), "\n ")
			} else {
				input = args[0]
			}

			return browser.OpenURL(input)
		},
	}
}
