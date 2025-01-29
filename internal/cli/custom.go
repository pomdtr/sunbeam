package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/history"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
)

func NewCmdCustom(alias string, extension extensions.Extension) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:     alias,
		Short:   extension.Manifest.Title,
		Long:    extension.Manifest.Description,
		Args:    cobra.NoArgs,
		GroupID: CommandGroupExtension,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isatty.IsTerminal(os.Stdout.Fd()) {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				encoder.SetEscapeHTML(false)

				return encoder.Encode(extension.Manifest)
			}

			if len(extension.Manifest.Root) == 0 {
				return cmd.Usage()
			}

			history, err := history.Load(history.Path)
			if err != nil {
				return err
			}

			rootList := tui.NewRootList(extension.Manifest.Title, history, func() ([]sunbeam.ListItem, error) {
				return extension.RootItems(), nil
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
		Short: command.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := make(map[string]any)

			for _, param := range command.Params {
				if !cmd.Flags().Changed(param.Name) {
					continue
				}

				switch param.Type {
				case sunbeam.InputString:
					value, err := cmd.Flags().GetString(param.Name)
					if err != nil {
						return err
					}
					payload[param.Name] = value
				case sunbeam.InputBoolean:
					value, err := cmd.Flags().GetBool(param.Name)
					if err != nil {
						return err
					}
					payload[param.Name] = value
				case sunbeam.InputNumber:
					value, err := cmd.Flags().GetInt(param.Name)
					if err != nil {
						return err
					}
					payload[param.Name] = value
				}
			}

			return runExtension(extension, command, payload)
		},
	}

	for _, input := range command.Params {
		switch input.Type {
		case sunbeam.InputString:
			cmd.Flags().String(input.Name, "", input.Description)
		case sunbeam.InputBoolean:
			cmd.Flags().Bool(input.Name, false, input.Description)
		case sunbeam.InputNumber:
			cmd.Flags().Int(input.Name, 0, input.Description)
		}

		if !input.Optional {
			_ = cmd.MarkFlagRequired(input.Name)
		}
	}

	return cmd
}

func runExtension(extension extensions.Extension, command sunbeam.Command, payload sunbeam.Payload) error {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		cmd, err := extension.CmdContext(context.Background(), command, payload)
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
		runner := tui.NewRunner(extension, command, payload)
		return tui.Draw(runner)
	case sunbeam.CommandModeSilent:
		cmd, err := extension.CmdContext(context.Background(), command, payload)
		if err != nil {
			return err
		}

		return cmd.Run()
	default:
		return fmt.Errorf("unknown command mode: %s", command.Mode)
	}
}
