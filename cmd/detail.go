package cmd

import (
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/sunbeamlauncher/sunbeam/app"
	"github.com/sunbeamlauncher/sunbeam/tui"
)

func NewCmdDetail() *cobra.Command {
	return &cobra.Command{
		Use:   "detail",
		Short: "Show detail view",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			var detail app.Detail
			err = json.Unmarshal(input, &detail)
			if err != nil {
				return err
			}

			view := tui.NewDetail("Filter")
			view.SetDetail(detail)
			model := tui.NewModel(view)

			return tui.Draw(model)
		},
	}
}
