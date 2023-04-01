package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mattn/go-isatty"
	"github.com/pkg/browser"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/pomdtr/sunbeam/utils"
	"github.com/spf13/cobra"
)

func NewTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger <action>",
		Short: "Trigger an action",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
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

			var generator internal.PageGenerator
			action = internal.ExpandAction(action, inputs)
			switch action.Type {
			case types.ReadAction:
				generator = internal.NewFileGenerator(action.Path)
			case types.RunAction:
				if action.OnSuccess != types.PushOnSuccess {
					if _, err := utils.RunCommand(action.Command, action.Dir); err != nil {
						exitWithErrorMsg("Unable to run command")
					}
					return
				}

				generator = internal.NewCommandGenerator(action.Command, action.Dir)
			case types.FetchAction:
				generator = internal.NewHttpGenerator(action.Url, action.Method, action.Headers, action.Body)
			case types.OpenAction:
				err := browser.OpenURL(args[0])
				if err != nil {
					fmt.Fprintln(os.Stderr, "Unable to open link:", err)
					os.Exit(1)
				}
				return
			case types.CopyAction:
				if err := clipboard.WriteAll(action.Text); err != nil {
					fmt.Fprintln(os.Stderr, "Unabble to write to clipboard:", err)
					os.Exit(1)
				}
				return
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Println(string(output))
				return
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()
		},
	}

	cmd.Flags().StringArrayP("inputs", "", nil, "inputs to pass to the action")

	return cmd
}
