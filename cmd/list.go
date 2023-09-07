package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Args:    cobra.NoArgs,
		GroupID: coreGroupID,
		Short:   "Parse items from stdin",
		RunE: func(cmd *cobra.Command, args []string) error {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no input provided")
			}

			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("could not read input: %s", err)
			}

			b = bytes.TrimSpace(b)

			var rows, lines [][]byte
			lines = bytes.Split(b, []byte("\n"))
			rows = make([][]byte, len(lines))
			for index, line := range lines {
				rows[index] = bytes.TrimRightFunc(line, func(r rune) bool {
					return r == '\r' || r == '\n'
				})
			}

			if len(rows) == 0 {
				return fmt.Errorf("no rows in input")
			}

			title := "Sunbeam"
			titleRow, _ := cmd.Flags().GetBool("title-row")
			if titleRow {
				title = string(rows[0])
				rows = rows[1:]
			}
			if cmd.Flags().Changed("title") {
				title, _ = cmd.Flags().GetString("title")
			}

			listItems := make([]types.ListItem, 0)
			delimiter, _ := cmd.Flags().GetString("delimiter")
			jsonInput, _ := cmd.Flags().GetBool("json")
			for _, row := range rows {
				if jsonInput {
					var v types.ListItem
					if err := json.Unmarshal(row, &v); err != nil {
						return fmt.Errorf("invalid JSON: %s", err)
					}
					listItems = append(listItems, v)
					continue
				}

				row := string(row)
				tokens := strings.Split(row, delimiter)

				// filter out empty tokens
				tokens = func(tokens []string) []string {
					var out []string
					for _, token := range tokens {
						if token != "" {
							out = append(out, token)
						}
					}
					return out
				}(tokens)

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
							Type:  types.ShareAction,
							Title: "Pipe",
							Text:  row,
						},
					},
				})
			}

			page := &types.List{
				Title: title,
				Items: listItems,
			}

			if err := json.NewEncoder(os.Stdout).Encode(page); err != nil {
				return fmt.Errorf("could not encode output: %s", err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("delimiter", "d", "\t", "delimiter")
	cmd.Flags().Bool("json", false, "json input")
	cmd.Flags().IntSlice("with-nth", nil, "indexes to show")
	cmd.MarkFlagsMutuallyExclusive("json", "delimiter")
	cmd.MarkFlagsMutuallyExclusive("json", "with-nth")

	cmd.Flags().String("title", "", "title")
	cmd.Flags().Bool("title-row", false, "use first row as title")
	cmd.MarkFlagsMutuallyExclusive("title", "title-row")

	cmd.Flags().Bool("show-detail", false, "show detail")
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
