package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/history"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/types"
	"github.com/spf13/cobra"
)

func NewCmdCustom(alias string, extension extensions.Extension, extensionConfig config.ExtensionConfig) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:     alias,
		Short:   extension.Manifest.Title,
		Long:    extension.Manifest.Description,
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

			if len(inputBytes) == 0 {
				if len(extension.RootItems()) == 0 && len(extensionConfig.Root) == 0 && len(extensionConfig.Items) == 0 {
					return cmd.Usage()
				}

				history, err := history.Load(history.Path)
				if err != nil {
					return err
				}

				rootList := tui.NewRootList(extension.Manifest.Title, history, func() (config.Config, []types.ListItem, error) {
					cfg, err := config.Load(config.Path)
					if err != nil {
						return config.Config{}, nil, err
					}

					items := extensionListItems(alias, extension, extensionConfig)
					return cfg, items, nil
				})

				return tui.Draw(rootList)
			}

			var input types.Payload
			if err := json.Unmarshal(inputBytes, &input); err != nil {
				return err
			}
			if input.Preferences == nil {
				input.Preferences = extensionConfig.Preferences
			}

			return runExtension(extension, input)
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	commands := extension.Manifest.Commands
	sort.Slice(extension.Manifest.Commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	for _, command := range commands {
		cmd := NewSubCmdCustom(alias, extension, extensionConfig, command)
		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

func NewSubCmdCustom(alias string, extension extensions.Extension, extensionConfig config.ExtensionConfig, command types.CommandSpec) *cobra.Command {
	parts := strings.Split(command.Name, ".")
	use := parts[len(parts)-1]
	cmd := &cobra.Command{
		Use:    use,
		Short:  command.Title,
		Hidden: command.Hidden,
		RunE: func(cmd *cobra.Command, args []string) error {
			params := make(map[string]any)

			for _, param := range command.Params {
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

			preferences := extensionConfig.Preferences
			if preferences == nil {
				preferences = make(map[string]any)
			}

			envs, err := tui.ExtractPreferencesFromEnv(alias, extension)
			if err != nil {
				return err
			}

			for name, value := range envs {
				preferences[name] = value
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

			return runExtension(extension, input)
		},
	}

	for _, input := range command.Params {
		switch input.Type {
		case types.InputText, types.InputTextArea, types.InputPassword:
			cmd.Flags().String(input.Name, "", input.Title)
		case types.InputCheckbox:
			cmd.Flags().Bool(input.Name, false, input.Title)
		case types.InputNumber:
			cmd.Flags().Int(input.Name, 0, input.Title)
		}

		if !input.Optional {
			_ = cmd.MarkFlagRequired(input.Name)
		}
	}

	return cmd
}

func runExtension(extension extensions.Extension, input types.Payload) error {
	command, ok := extension.Command(input.Command)
	if !ok {
		return fmt.Errorf("command %s not found", input.Command)
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
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
	case types.CommandModeSearch, types.CommandModeFilter, types.CommandModeDetail:
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
