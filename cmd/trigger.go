package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/pkg/browser"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
	"github.com/spf13/cobra"
)

func NewTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "trigger",
		Args:    cobra.NoArgs,
		GroupID: coreGroupID,
		Short:   "Trigger an action",
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

			if len(inputsFlag) < len(action.Inputs) && !isatty.IsTerminal(os.Stdout.Fd()) {
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

	cmd.Flags().StringArrayP("input", "i", nil, "input values")
	cmd.Flags().StringP("query", "q", "", "query value")

	return cmd
}

func triggerAction(action types.Action, inputs map[string]string, query string) error {
	for name, value := range inputs {
		action = internal.RenderAction(action, fmt.Sprintf("${input:%s}", name), value)
	}

	action = internal.RenderAction(action, "${query}", query)

	if len(inputs) < len(action.Inputs) {
		missing := make([]types.Input, 0)
		for _, input := range action.Inputs {
			if _, ok := inputs[input.Name]; !ok {
				missing = append(missing, input)
			}
		}

		return internal.Draw(internal.NewForm(action.Title, func(values map[string]string) tea.Msg {
			action := action
			for name, value := range values {
				action = internal.RenderAction(action, fmt.Sprintf("${input:%s}", name), value)
			}

			switch action.Type {
			case types.PushAction:
				var page internal.Page

				if action.Page == nil {
					return fmt.Errorf("no page provided")
				}

				if action.Page.Command != nil {
					page = internal.NewRunner(internal.NewCommandGenerator(action.Command))
				} else {
					page = internal.NewRunner(internal.NewFileGenerator(action.Page.Text))
				}

				return internal.PushPageMsg{
					Page: page,
				}
			case types.RunAction:
				return internal.ExitMsg{
					Cmd: action.Command.Cmd(),
				}
			case types.OpenAction:
				err := browser.OpenURL(action.Target)
				if err != nil {
					return fmt.Errorf("unable to open link: %s", err)
				}
				return tea.Quit()
			case types.CopyAction:
				if err := clipboard.WriteAll(action.Text); err != nil {
					return fmt.Errorf("unable to write to clipboard: %s", err)
				}
				return tea.Quit()

			default:
				return fmt.Errorf("unknown action type: %s", action.Type)
			}
		}, missing...), options)
	}

	switch action.Type {
	case types.PushAction:
		if action.Page.Command != nil {
			return Run(internal.NewCommandGenerator(action.Page.Command))
		}
		return Run(internal.NewFileGenerator(action.Page.Text))
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
