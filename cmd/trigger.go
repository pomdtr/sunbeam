package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger <action>",
		Short: "Trigger an action",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var generator tui.PageGenerator
			inputsFlag, _ := cmd.Flags().GetStringArray("inputs")
			inputs := make(map[string]string)
			for _, input := range inputsFlag {
				parts := strings.SplitN(input, "=", 2)
				if len(parts) != 2 {
					exitWithErrorMsg("invalid input: %s", input)
				}
				inputs[parts[0]] = parts[1]
			}

			if isatty.IsTerminal(os.Stdin.Fd()) {
				exitWithErrorMsg("No input provided")
			}

			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				exitWithErrorMsg("could not read input: %s", err)
			}

			var action types.Action
			if err := json.Unmarshal(input, &action); err != nil {
				exitWithErrorMsg("could not parse input: %s", err)
			}

			if action.Type == types.CopyAction {
				err := clipboard.WriteAll(action.Text)
				if err != nil {
					exitWithErrorMsg("could not copy to clipboard: %s", err)
				}
				return
			}

			generator = tui.NewActionGenerator(action, inputs)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Println(string(output))
				return
			}

			runner := tui.NewRunner(generator)
			tui.NewPaginator(runner).Draw()
		},
	}

	cmd.Flags().StringArrayP("inputs", "", nil, "inputs to pass to the action")

	return cmd
}
