package cmd

import (
	"io"
	"os"
	"strconv"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewCmdFilter() *cobra.Command {
	return &cobra.Command{
		Use:   "filter",
		Short: "Show filter",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			items, err := app.ParseListItems(string(input))
			if err != nil {
				return err
			}

			listItems := make([]tui.ListItem, len(items))
			for i, item := range items {
				item.Id = strconv.Itoa(i)
				listItems[i] = tui.ParseScriptItem(item)
			}

			list := tui.NewList("Filter")
			list.SetItems(listItems)
			model := tui.NewModel(list)

			return tui.Draw(model)
		},
	}
}
