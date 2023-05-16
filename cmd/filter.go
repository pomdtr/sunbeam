package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewFilterCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "filter",
		Args:    cobra.NoArgs,
		GroupID: coreGroupID,
		Short:   "Filter stdin rows",
		RunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no input provided")
			}
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("could not read input: %s", err)
			}

			input := string(b)
			rows := strings.Split(input, "\n")

			return Run(func() (*types.Page, error) {
				listItems := make([]types.ListItem, len(rows))
				for i, row := range rows {
					listItems[i] = types.ListItem{
						Title: row,
						Actions: []types.Action{
							{
								Type:  types.PasteAction,
								Title: "Paste",
								Text:  row,
							},
						},
					}
				}

				return &types.Page{
					Type:  types.ListPage,
					Title: "Filter",
					Items: listItems,
				}, nil
			})

		},
	}
}
