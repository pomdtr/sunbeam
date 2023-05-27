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

			var rows []string
			if runtime.GOOS == "windows" {
				rows = strings.Split(string(b), "\r\n")
			} else {
				rows = strings.Split(string(b), "\n")
			}

			if len(rows) == 0 {
				return fmt.Errorf("now rows in input")
			}

			var header string
			headerLine, _ := cmd.Flags().GetBool("header-line")
			if headerLine {
				header = rows[0]
				rows = rows[1:]
			} else {
				header = "Sunbeam"
			}

			return Run(func() (*types.Page, error) {
				listItems := make([]types.ListItem, 0)
				for _, row := range rows {
					if row == "" {
						continue
					}
					delimiter, _ := cmd.Flags().GetString("delimiter")
					tokens := strings.Split(row, delimiter)

					var title, subtitle string
					var accessories []string
					if cmd.Flags().Changed("with-nth") {
						nths, _ := cmd.Flags().GetIntSlice("with-nth")
						title = safeGet(tokens, nths[0])
						if len(nths) > 1 {
							subtitle = safeGet(tokens, nths[1])
						}
						if len(nths) > 2 {
							for _, nth := range nths[2:] {
								accessories = append(accessories, safeGet(tokens, nth))
							}
						}
					} else {
						title = tokens[0]
						if len(tokens) > 1 {
							subtitle = tokens[1]
						}
						if len(tokens) > 2 {
							accessories = tokens[2:]
						}
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
					Title: header,
					Items: listItems,
				}, nil
			})

		},
	}

	cmd.Flags().StringP("delimiter", "d", "\t", "delimiter")
	cmd.Flags().IntSlice("with-nth", nil, "indexes to show")
	cmd.Flags().BoolP("header-line", "H", false, "treat the first line as the page title")
	return cmd
}

func safeGet(tokens []string, idx int) string {
	if idx == 0 {
		return ""
	}
	if idx > len(tokens) {
		return ""
	}

	return tokens[idx-1]
}
