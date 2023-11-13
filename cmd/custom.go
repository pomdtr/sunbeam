package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
)

func NewCmdCustom(alias string, extension extensions.Extension, preferences map[string]any) (*cobra.Command, error) {

	rootCmd := &cobra.Command{
		Use:     alias,
		Short:   extension.Title,
		Long:    extension.Description,
		Args:    cobra.NoArgs,
		GroupID: CommandGroupExtension,
		RunE: func(cmd *cobra.Command, args []string) error {
			var inputBytes []byte
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				inputBytes = b
			}

			if len(inputBytes) > 0 {
				var input types.Payload
				if err := json.Unmarshal(inputBytes, &input); err != nil {
					return err
				}
				if input.Preferences == nil {
					input.Preferences = preferences
				}

				var rawOutput bool
				if cmd.Flags().Changed("raw") {
					rawOutput, _ = cmd.Flags().GetBool("raw")
				} else {
					rawOutput = !isatty.IsTerminal(os.Stdout.Fd())
				}

				return runExtension(extension, input, rawOutput)
			}

			return cmd.Usage()
		},
	}

	rootCmd.Flags().Bool("raw", false, "raw output")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	for _, command := range extension.Commands {
		command := command
		cmd := &cobra.Command{
			Use:    command.Name,
			Short:  command.Title,
			Hidden: command.Hidden,
			RunE: func(cmd *cobra.Command, args []string) error {
				params := make(map[string]any)

				for _, param := range command.Inputs {
					if !cmd.Flags().Changed(param.Name) {
						continue
					}

					switch param.Type {
					case types.InputText, types.InputTextArea, types.InputPassword:
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}
						params[param.Name] = value
					case types.InputCheckbox:
						value, err := cmd.Flags().GetBool(param.Name)
						if err != nil {
							return err
						}
						params[param.Name] = value
					case types.InputNumber:
						value, err := cmd.Flags().GetInt(param.Name)
						if err != nil {
							return err
						}
						params[param.Name] = value
					}
				}

				input := types.Payload{
					Command:     command.Name,
					Preferences: preferences,
					Params:      params,
				}

				if !isatty.IsTerminal(os.Stdin.Fd()) {
					stdin, err := io.ReadAll(os.Stdin)
					if err != nil {
						return err
					}

					input.Query = string(bytes.Trim(stdin, "\n"))
				}

				return runExtension(extension, input, !isatty.IsTerminal(os.Stdout.Fd()))
			},
		}

		for _, input := range command.Inputs {
			switch input.Type {
			case types.InputText, types.InputTextArea, types.InputPassword:
				cmd.Flags().String(input.Name, "", input.Title)
			case types.InputCheckbox:
				cmd.Flags().Bool(input.Name, false, input.Title)
			case types.InputNumber:
				cmd.Flags().Int(input.Name, 0, input.Title)
			}

			if input.Required {
				_ = cmd.MarkFlagRequired(input.Name)
			}
		}

		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

func runExtension(extension extensions.Extension, input types.Payload, rawOutput bool) error {
	if err := extension.CheckRequirements(); err != nil {
		return tui.Draw(tui.NewErrorPage(fmt.Errorf("missing requirements: %w", err)))
	}

	command, ok := extension.Command(input.Command)
	if !ok {
		return fmt.Errorf("command %s not found", input.Command)
	}

	if missing := tui.FindMissingInputs(command.Inputs, input.Params); len(missing) > 0 {
		names := make([]string, len(missing))
		for i, param := range missing {
			names[i] = param.Name
		}
		return fmt.Errorf("missing required params: %s", strings.Join(names, ", "))
	}

	if rawOutput {
		cmd, err := extension.Cmd(input)
		if err != nil {
			return err
		}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}

	switch command.Mode {
	case types.CommandModeList, types.CommandModeDetail:
		runner := tui.NewRunner(extension, input)
		return tui.Draw(runner)
	case types.CommandModeSilent:
		return extension.Run(input)
	case types.CommandModeTTY:
		cmd, err := extension.Cmd(input)
		if err != nil {
			return err
		}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	default:
		return fmt.Errorf("unknown command mode: %s", command.Mode)
	}
}
