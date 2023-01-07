package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

func NewCmdOpen() *cobra.Command {
	return &cobra.Command{
		Use:     "open",
		Short:   "Open file with default app",
		GroupID: "core",
		Args:    cobra.MaximumNArgs(1),
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

			return open.Run(input)
		},
	}
}
