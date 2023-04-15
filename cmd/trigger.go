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
	"github.com/spf13/cobra"
)

func NewTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger <action>",
		Short: "Trigger an action",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			inputsFlag, _ := cmd.Flags().GetStringArray("inputs")
			if len(inputsFlag) < len(action.Inputs) {
				return fmt.Errorf("not enough inputs provided")
			}

			inputs := make(map[string]string)
			for _, input := range inputsFlag {
				parts := strings.SplitN(input, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid input: %s", input)
				}
				inputs[parts[0]] = parts[1]
			}

			for name, value := range inputs {
				action = internal.RenderAction(action, fmt.Sprintf("${input:%s}", name), value)
			}

			switch action.Type {
			case types.PushAction:
				return Draw(internal.NewFileGenerator(action.Page))
			case types.RunAction:
				if action.OnSuccess == types.PushOnSuccess {
					return Draw(internal.NewCommandGenerator(action.Command))
				}

				if _, err := action.Command.Output(); err != nil {
					return fmt.Errorf("command failed: %s", err)
				}

				switch action.OnSuccess {
				case types.ExitOnSuccess, types.ReloadOnSuccess:
					return nil
				case types.CopyOnSuccess:
					if err := clipboard.WriteAll(action.Text); err != nil {
						return fmt.Errorf("unable to write to clipboard: %s", err)
					}
					return nil
				case types.OpenOnSuccess:
					err := browser.OpenURL(action.Text)
					if err != nil {
						return fmt.Errorf("unable to open link: %s", err)
					}
					return nil
				default:
					return nil
				}

			case types.OpenAction:
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
