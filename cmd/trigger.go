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
		RunE: func(cmd *cobra.Command, args []string) error {
			inputsFlag, _ := cmd.Flags().GetStringArray("inputs")
			inputs := make(map[string]string)
			for _, input := range inputsFlag {
				parts := strings.SplitN(input, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid input: %s", input)
				}
				inputs[parts[0]] = parts[1]
			}

			if isatty.IsTerminal(os.Stdin.Fd()) {
				return fmt.Errorf("no input provided")
			}

			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("could not read input: %s", err)
			}

			var action types.Action
			if err := json.Unmarshal(input, &action); err != nil {
				return fmt.Errorf("could not parse input: %s", err)
			}

			if action.Type == types.CopyAction {
				err := clipboard.WriteAll(action.Text)
				if err != nil {
					return fmt.Errorf("could not copy to clipboard: %s", err)
				}
				return nil
			}

			for name, value := range inputs {
				action = internal.ExpandAction(action, fmt.Sprintf("${input:%s}", name), value)
			}

			switch action.Type {
			case types.PushPageAction:
				var generator internal.PageGenerator
				switch action.Page.Type {
				case types.StaticTarget:
					generator = internal.NewFileGenerator(action.Path)
				case types.DynamicTarget:
					generator = internal.NewCommandGenerator(action.Page.Command, action.Page.Input, action.Page.Dir)
				}
				if !isatty.IsTerminal(os.Stdout.Fd()) {
					output, err := generator()
					if err != nil {
						return fmt.Errorf("could not generate page: %s", err)
					}

					fmt.Println(string(output))
					return nil
				}

				runner := internal.NewRunner(generator)
				internal.NewPaginator(runner).Draw()
				return nil
			case types.RunAction:
				if _, err := utils.RunCommand(action.Command, strings.NewReader(action.Input), action.Dir); err != nil {
					return fmt.Errorf("unable to run command")
				}
				return nil
			case types.OpenPathAction, types.OpenUrlAction:
				err := browser.OpenURL(args[0])
				if err != nil {
					return fmt.Errorf("unable to open link: %s", err)
				}
				return nil
			case types.CopyAction:
				if err := clipboard.WriteAll(action.Text); err != nil {
					return fmt.Errorf("unable to write to clipboard: %s", err)
				}
				return nil
			default:
				return fmt.Errorf("unknown action type: %s", action.Type)
			}
		},
	}

	cmd.Flags().StringArrayP("inputs", "", nil, "inputs to pass to the action")

	return cmd
}
