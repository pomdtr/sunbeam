package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
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
			cfg, err := config.Load()
			if err != nil {
				return err
			}

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

				return runExtension(extension, input, cfg.Env)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(extension)
			}

			if len(extension.RootItems()) == 0 {
				return cmd.Usage()
			}

			page := tui.NewRootList(extension.Title, func() (extensions.ExtensionMap, []types.ListItem, map[string]string, error) {
				cfg, err := config.Load()
				if err != nil {
					return nil, nil, nil, err
				}

				extension, err := LoadExtension(extension.Entrypoint)
				if err != nil {
					return nil, nil, nil, err
				}

				extensionMap := extensions.ExtensionMap{alias: extension}
				return extensionMap, ExtensionRootItems(alias, extension), cfg.Env, nil
			})
			return tui.Draw(page)
		},
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	for _, command := range extension.Commands {
		command := command
		cmd := &cobra.Command{
			Use:    command.Name,
			Short:  command.Title,
			Hidden: command.Hidden,
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := config.Load()
				if err != nil {
					return err
				}

				params := make(map[string]any)

				for _, param := range command.Params {
					if !cmd.Flags().Changed(param.Name) {
						continue
					}

					switch param.Type {
					case types.ParamTypeString:
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}
						params[param.Name] = value
					case types.ParamTypeBoolean:
						value, err := cmd.Flags().GetBool(param.Name)
						if err != nil {
							return err
						}
						params[param.Name] = value
					case types.ParamTypeNumber:
						value, err := cmd.Flags().GetInt(param.Name)
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

					input.Query = string(stdin)
				}

				return runExtension(extension, input, cfg.Env)
			},
		}

		for _, param := range command.Params {
			switch param.Type {
			case types.ParamTypeString:
				cmd.Flags().String(param.Name, "", param.Description)
			case types.ParamTypeBoolean:
				cmd.Flags().Bool(param.Name, false, param.Description)
			case types.ParamTypeNumber:
				cmd.Flags().Int(param.Name, 0, param.Description)
			}
		}

		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

func runExtension(extension extensions.Extension, input types.CommandInput, environ map[string]string) error {
	if err := extension.CheckRequirements(); err != nil {
		return tui.Draw(tui.NewErrorPage(fmt.Errorf("missing requirements: %w", err)))
	}

	command, ok := extension.Command(input.Command)
	if !ok {
		return fmt.Errorf("command %s not found", input.Command)
	}

	missing := tui.FindMissingParams(command.Params, input.Params)
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		if len(missing) > 0 {
			names := make([]string, len(missing))
			for i, param := range missing {
				names[i] = param.Name
			}
			return fmt.Errorf("missing required params: %s", strings.Join(names, ", "))
		}
		output, err := extension.Output(input, environ)

		if err != nil {
			return err
		}

		if _, err := os.Stdout.Write(output); err != nil {
			return err
		}

		return nil
	}

	if len(missing) > 0 {
		cancelled := true
		form := tui.NewForm(func(values map[string]any) tea.Msg {
			cancelled = false
			for k, v := range values {
				input.Params[k] = v
			}

			return tui.ExitMsg{}
		}, missing...)

		if err := tui.Draw(form); err != nil {
			return err
		}

		if cancelled {
			return nil
		}
	}

	switch command.Mode {
	case types.CommandModeList, types.CommandModeDetail:
		runner := tui.NewRunner(extension, input, environ)
		return tui.Draw(runner)
	case types.CommandModeSilent:
		return extension.Run(input, environ)
	case types.CommandModeTTY:
		cmd, err := extension.Cmd(input, environ)
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
