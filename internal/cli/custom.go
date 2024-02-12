package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
)

func NewCmdCustom(alias string, extension extensions.Extension, cfg config.Config) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   alias,
		Short: extension.Manifest.Title,
		Long:  extension.Manifest.Description,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isatty.IsTerminal(os.Stdout.Fd()) {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				encoder.SetEscapeHTML(false)

				return encoder.Encode(extension.Manifest)
			}

			var actions []sunbeam.Action
			actions = append(actions, extension.Manifest.Root...)
			for _, action := range cfg.Root {
				if action.Extension == alias {
					actions = append(actions, action)
				}
			}
			if len(actions) == 0 {
				return cmd.Help()
			}

			rootList := tui.NewRootList(extension.Manifest.Title, nil, func() (config.Config, []sunbeam.ListItem, error) {
				var items []sunbeam.ListItem
				for _, action := range actions {
					action.Extension = alias
					if action.Title == "" {
						command, ok := extension.Command(action.Command)
						if !ok {
							continue
						}
						action.Title = command.Title
					}
					items = append(items, sunbeam.ListItem{
						Title:       action.Title,
						Accessories: []string{alias},
						Actions:     []sunbeam.Action{action},
					})
				}

				return cfg, items, nil
			})

			return tui.Draw(rootList)
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	commands := extension.Manifest.Commands
	sort.Slice(extension.Manifest.Commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	for _, command := range commands {
		cmd := NewSubCmdCustom(alias, extension, command)
		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

func NewSubCmdCustom(alias string, extension extensions.Extension, command sunbeam.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   command.Name,
		Short: command.Title,
		RunE: func(cmd *cobra.Command, args []string) error {
			params := make(map[string]any)

			for _, param := range command.Params {
				if !cmd.Flags().Changed(param.Name) {
					continue
				}

				switch param.Type {
				case sunbeam.ParamString:
					value, err := cmd.Flags().GetString(param.Name)
					if err != nil {
						return err
					}
					params[param.Name] = value
				case sunbeam.ParamBoolean:
					value, err := cmd.Flags().GetBool(param.Name)
					if err != nil {
						return err
					}
					params[param.Name] = value
				case sunbeam.ParamNumber:
					value, err := cmd.Flags().GetInt(param.Name)
					if err != nil {
						return err
					}
					params[param.Name] = value
				}
			}

			input := sunbeam.Payload{
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

			return runExtension(extension, input)
		},
	}

	if command.Hidden {
		cmd.Hidden = true
	}

	for _, input := range command.Params {
		switch input.Type {
		case sunbeam.ParamString:
			cmd.Flags().String(input.Name, "", input.Title)
		case sunbeam.ParamBoolean:
			cmd.Flags().Bool(input.Name, false, input.Title)
		case sunbeam.ParamNumber:
			cmd.Flags().Int(input.Name, 0, input.Title)
		}

		if !input.Optional {
			_ = cmd.MarkFlagRequired(input.Name)
		}
	}

	return cmd
}

func runExtension(extension extensions.Extension, input sunbeam.Payload) error {
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
	case sunbeam.CommandModeSearch, sunbeam.CommandModeFilter, sunbeam.CommandModeDetail:
		runner := tui.NewRunner(extension, input)
		return tui.Draw(runner)
	case sunbeam.CommandModeSilent:
		return extension.Run(input)
	case sunbeam.CommandModeTTY:
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
