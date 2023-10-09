package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"muzzammil.xyz/jsonc"
)

func NewCmdCustom(extensions map[string]tui.Extension, alias string) (*cobra.Command, error) {
	extension, ok := extensions[alias]
	if !ok {
		return nil, fmt.Errorf("extension %s not found", alias)
	}
	rootCmd := &cobra.Command{
		Use:     alias,
		Short:   extension.Title,
		Long:    extension.Description,
		Args:    cobra.NoArgs,
		GroupID: CommandGroupExtension,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isatty.IsTerminal(os.Stdout.Fd()) {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(extension.Manifest)
			}

			rootCommands := make([]types.CommandSpec, 0)
			for _, command := range extension.Commands {
				if !IsRootCommand(command) {
					continue
				}

				rootCommands = append(rootCommands, command)
			}

			if len(rootCommands) == 0 {
				return cmd.Usage()
			}

			if len(rootCommands) == 1 {
				command := rootCommands[0]
				runner, err := tui.NewRunner(extensions, types.CommandRef{
					Extension: alias,
					Command:   command.Name,
				})

				if err != nil {
					return err
				}
				return tui.Draw(runner, MaxHeight)
			}

			page := tui.NewRootList(extension.Title, func() (map[string]tui.Extension, []types.ListItem, error) {
				items := make([]types.ListItem, 0)
				for _, command := range rootCommands {
					items = append(items, types.ListItem{
						Title:    command.Title,
						Subtitle: extension.Title,
						Actions: []types.Action{
							{
								Title: "Run Command",
								OnAction: types.Command{
									Type: types.CommandTypeRun,
									CommandRef: types.CommandRef{
										Extension: alias,
										Command:   command.Name,
									},
								},
							},
						},
					})
				}
				return extensions, items, nil
			})

			return tui.Draw(page, MaxHeight)
		},
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	for _, subcommand := range extension.Commands {
		subcommand := subcommand
		cmd := &cobra.Command{
			Use:    subcommand.Name,
			Short:  subcommand.Title,
			Hidden: subcommand.Hidden,
			RunE: func(cmd *cobra.Command, args []string) error {
				params := make(map[string]any)
				for _, param := range subcommand.Params {
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

				if subcommand.Mode == types.CommandModeView {
					if !isatty.IsTerminal(os.Stdout.Fd()) {
						output, err := extension.Run(subcommand.Name, types.CommandInput{
							Params: params,
						})

						if err != nil {
							return err
						}

						if _, err := os.Stdout.Write(output); err != nil {
							return err
						}

						return nil
					}

					runner, err := tui.NewRunner(extensions, types.CommandRef{
						Extension: alias,
						Command:   subcommand.Name,
						Params:    params,
					})

					if err != nil {
						return err
					}

					return tui.Draw(runner, MaxHeight)
				}

				out, err := extension.Run(subcommand.Name, types.CommandInput{
					Params: params,
				})
				if err != nil {
					return err
				}

				if len(out) == 0 {
					return nil
				}

				var command types.Command
				if err := jsonc.Unmarshal(out, &command); err != nil {
					return err
				}

				switch command.Type {
				case types.CommandTypeCopy:
					return clipboard.WriteAll(command.Text)
				case types.CommandTypeOpen:
					return utils.OpenWith(command.Target, command.App)
				default:
					return nil
				}
			},
		}

		for _, param := range subcommand.Params {
			switch param.Type {
			case types.ParamTypeString:
				cmd.Flags().String(param.Name, "", param.Description)
			case types.ParamTypeBoolean:
				cmd.Flags().Bool(param.Name, false, param.Description)
			case types.ParamTypeNumber:
				cmd.Flags().Int(param.Name, 0, param.Description)
			}

			if !param.Optional {
				_ = cmd.MarkFlagRequired(param.Name)
			}
		}

		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

func buildDoc(command *cobra.Command) (string, error) {
	var page strings.Builder
	err := doc.GenMarkdown(command, &page)
	if err != nil {
		return "", err
	}

	out := strings.Builder{}
	for _, line := range strings.Split(page.String(), "\n") {
		if strings.Contains(line, "SEE ALSO") {
			break
		}

		out.WriteString(line + "\n")
	}

	for _, child := range command.Commands() {
		childPage, err := buildDoc(child)
		if err != nil {
			return "", err
		}
		out.WriteString(childPage)
	}

	return out.String(), nil
}

func LookupIntEnv(key string, fallback int) int {
	env, ok := os.LookupEnv(key)
	if !ok {
		return fallback

	}

	value, err := strconv.Atoi(env)
	if err != nil {
		return fallback
	}

	return value
}

func LookupBoolEnv(key string, fallback bool) bool {
	env, ok := os.LookupEnv(key)
	if !ok {
		return fallback

	}

	b, err := strconv.ParseBool(env)
	if err != nil {
		return fallback
	}

	return b
}
