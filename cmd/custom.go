package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/schemas"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewCmdCustom(alias string, extension extensions.Extension) (*cobra.Command, error) {
	exts := extensions.ExtensionMap{
		alias: extension,
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

			rootCommands := extension.RootCommands()
			if len(rootCommands) == 0 {
				return cmd.Usage()
			}

			if len(rootCommands) == 1 {
				runner := tui.NewRunner(extension, rootCommands[0], types.CommandInput{
					Params: make(map[string]any),
				})

				return tui.Draw(runner)
			}

			page := tui.NewRootList(extension.Title, exts)

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
				input := types.CommandInput{
					Params: make(map[string]any),
				}

				// load params from stdin
				if !isatty.IsTerminal(os.Stdin.Fd()) {
					i, err := io.ReadAll(os.Stdin)
					if err != nil {
						return err
					}

					if len(i) > 0 {
						if err := json.Unmarshal(i, &input); err != nil {
							return err
						}
					}
				}

				for _, param := range command.Params {
					if !cmd.Flags().Changed(param.Name) {
						if _, ok := input.Params[param.Name]; ok {
							continue
						}

						if param.Required {
							return fmt.Errorf("%s is a required parameter", param.Name)
						}
					}

					switch param.Type {
					case types.ParamTypeString:
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}
						input.Params[param.Name] = value
					case types.ParamTypeBoolean:
						value, err := cmd.Flags().GetBool(param.Name)
						if err != nil {
							return err
						}
						input.Params[param.Name] = value
					}
				}

				if !isatty.IsTerminal(os.Stdout.Fd()) {
					output, err := extension.Run(command.Name, input)

					if err != nil {
						return err
					}

					if _, err := os.Stdout.Write(output); err != nil {
						return err
					}

					return nil
				}

				var runCommand func(command types.CommandSpec, input types.CommandInput) error
				runCommand = func(command types.CommandSpec, input types.CommandInput) error {
					if command.Mode == types.CommandModeView {
						runner := tui.NewRunner(extension, command, input)
						return tui.Draw(runner)
					}

					output, err := extension.Run(command.Name, input)
					if err != nil {
						return err
					}

					if len(output) == 0 {
						return nil
					}

					if err := schemas.ValidateCommand(output); err != nil {
						return err
					}

					var res types.Command
					if err := json.Unmarshal(output, &res); err != nil {
						return err
					}

					switch res.Type {
					case types.CommandTypeCopy:
						return clipboard.WriteAll(res.Text)
					case types.CommandTypeOpen:
						return utils.OpenWith(res.Target, res.App)
					case types.CommandTypeRun:
						target, ok := extension.Command(res.Command)
						if !ok {
							return fmt.Errorf("command %s not found", res.Command)
						}
						return runCommand(target, input)
					case types.CommandTypeExit, types.CommandTypeReload:
						return nil
					default:
						return fmt.Errorf("unknown command type %s", res.Type)
					}
				}

				return runCommand(command, input)
			},
		}

		for _, param := range command.Params {
			switch param.Type {
			case types.ParamTypeString:
				if param.Default != nil {
					defaultValue, ok := param.Default.(string)
					if !ok {
						return nil, fmt.Errorf("invalid default value for parameter %s", param.Name)
					}
					cmd.Flags().String(param.Name, defaultValue, param.Description)
				} else {
					cmd.Flags().String(param.Name, "", param.Description)
				}
			case types.ParamTypeBoolean:
				if param.Default != nil {
					defaultValue, ok := param.Default.(bool)
					if !ok {
						return nil, fmt.Errorf("invalid default value for parameter %s", param.Name)
					}

					cmd.Flags().Bool(param.Name, defaultValue, param.Description)
				} else {
					cmd.Flags().Bool(param.Name, false, param.Description)
				}
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
