package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/pkg/browser"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger",
		Args:  cobra.NoArgs,
		Short: "Trigger an action",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("could not read input: %s", err)
			}

			var action types.Action
			if err := json.Unmarshal(input, &action); err != nil {
				return fmt.Errorf("could not parse input: %s", err)
			}

			inputsFlag, _ := cmd.Flags().GetStringArray("input")
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

			query, _ := cmd.Flags().GetString("query")
			return triggerAction(action, inputs, query)
		},
	}

	cmd.Flags().StringArrayP("input", "i", nil, "Input values")
	cmd.Flags().StringP("query", "q", "", "Query")

	return cmd
}

func triggerAction(action types.Action, inputs map[string]string, query string) error {
	for name, value := range inputs {
		action = internal.RenderAction(action, fmt.Sprintf("${input:%s}", name), value)
	}

	action = internal.RenderAction(action, "${query}", query)

	switch action.Type {
	case types.PushAction:
		if action.Command != nil {
			return Draw(internal.NewCommandGenerator(action.Command))
		}
		return Draw(internal.NewFileGenerator(action.Page))
	case types.RunAction:
		if _, err := action.Command.Output(); err != nil {
			return fmt.Errorf("command failed: %s", err)
		}

		return nil
	case types.OpenAction:
		err := browser.OpenURL(action.Target)
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
}
