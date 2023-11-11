package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/extensions"
	preferences "github.com/pomdtr/sunbeam/internal/storage"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
)

func NewCmdCustom(alias string, extension extensions.Extension) (*cobra.Command, error) {

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
				var input types.CommandInput
				if err := json.Unmarshal(inputBytes, &input); err != nil {
					return err
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
					case types.InputTextField, types.InputPassword:
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
					}
				}

				input := types.CommandInput{
					Command: command.Name,
					Params:  params,
				}

				if !isatty.IsTerminal(os.Stdin.Fd()) {
					stdin, err := io.ReadAll(os.Stdin)
					if err != nil {
						return err
					}

					input.Query = string(bytes.Trim(stdin, "\n"))
				}

				prefs, err := preferences.Load(alias, extension.Origin)
				if err != nil {
					return err
				}
				input.Preferences = prefs

				if missing := tui.FindMissingInputs(extension.Preferences, input.Preferences); len(missing) > 0 {
					cancelled := true
					var formError error
					form := tui.NewForm(func(values map[string]any) tea.Msg {
						cancelled = false
						input.Preferences = values
						formError = preferences.Save(alias, extension.Origin, values)
						return tui.ExitMsg{}
					}, missing...)

					if err := tui.Draw(form); err != nil {
						return err
					}

					if cancelled {
						return nil
					}

					if formError != nil {
						return formError
					}
				}

				return runExtension(extension, input, !isatty.IsTerminal(os.Stdout.Fd()))
			},
		}

		for _, input := range command.Inputs {
			switch input.Type {
			case types.InputTextField, types.InputPassword:
				cmd.Flags().String(input.Name, "", input.Title)
			case types.InputCheckbox:
				cmd.Flags().Bool(input.Name, false, input.Title)
			}

			if input.Required {
				_ = cmd.MarkFlagRequired(input.Name)
			}
		}

		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

func runExtension(extension extensions.Extension, input types.CommandInput, rawOutput bool) error {
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
