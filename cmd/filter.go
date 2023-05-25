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
	cmd := &cobra.Command{
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
				listItems := make([]types.ListItem, 0)
				for _, row := range rows {
					if row == "" {
						continue
					}

					delimiter, _ := cmd.Flags().GetString("delimiter")
					tokens := strings.Split(row, delimiter)
					var title string
					var subtitle string
					accessories := make([]string, 0)
					if len(tokens) == 1 {
						title = tokens[0]
					} else if len(tokens) == 2 {
						title = tokens[0]
						subtitle = tokens[1]
					} else {
						title = tokens[0]
						subtitle = tokens[1]
						accessories = tokens[2:]
					}

					listItems = append(listItems, types.ListItem{
						Title:       title,
						Subtitle:    subtitle,
						Accessories: accessories,
						Actions: []types.Action{
							{
								Type:  types.PasteAction,
								Title: "Confirm",
								Text:  row,
							},
						},
					})
				}

				return &types.Page{
					Type:  types.ListPage,
					Title: "Filter",
					Items: listItems,
				}, nil
			})

		},
	}

	cmd.Flags().StringP("delimiter", "d", "\t", "delimiter")
	return cmd
}
