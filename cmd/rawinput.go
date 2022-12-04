package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func NewRawInputCommand(config tui.Config) *cobra.Command {
	rawInputCmd := &cobra.Command{
		Use: "read",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			format, err := cmd.Flags().GetString("format")
			if err != nil {
				return err
			}
			title, err := cmd.Flags().GetString("title")
			if err != nil {
				return err
			}
			if title == "" {
				title = cases.Title(language.AmericanEnglish).String(format)
			}

			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			var page tui.Container
			switch format {
			case "list":
				scriptItems, err := app.ParseListItems(string(bytes))
				if err != nil {
					return err
				}
				listItems := make([]tui.ListItem, len(scriptItems))
				for i, scriptItem := range scriptItems {
					listItems[i] = tui.ParseScriptItem(scriptItem)
				}

				list := tui.NewList(title)
				list.SetItems(listItems)
				page = list
			case "detail":
				detail := tui.NewDetail(title)
				detail.SetContent(string(bytes))
				page = detail
			default:
				return fmt.Errorf("unknown format %s", format)
			}

			return tui.Draw(page, config)
		},
	}

	rawInputCmd.Flags().StringP("title", "t", "", "Title of the list")
	rawInputCmd.Flags().StringP("format", "f", "list", "Format of the input")
	rawInputCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"list", "detail"}, cobra.ShellCompDirectiveNoFileComp
	})
	rawInputCmd.MarkFlagRequired("format")

	return rawInputCmd
}
