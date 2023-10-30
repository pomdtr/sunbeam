package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewCmdCustom(extension extensions.Extension) (*cobra.Command, error) {
	exts := extensions.ExtensionMap{
		extension.Alias: extension,
	}
	rootCmd := &cobra.Command{
		Use:     extension.Alias,
		Short:   extension.Title,
		Long:    extension.Description,
		Args:    cobra.NoArgs,
		GroupID: CommandGroupExtension,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isatty.IsTerminal(os.Stdout.Fd()) {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(extension)
			}

			page := tui.NewRootList(extension.Title, func() (extensions.ExtensionMap, []types.RootItem, error) {
				extensions, err := FindExtensions()
				if err != nil {
					return nil, nil, err
				}

				extension, ok := extensions[extension.Alias]
				if !ok {
					return nil, nil, fmt.Errorf("extension not found: %s", extension.Alias)
				}

				var rootItems []types.RootItem
				rootItems = append(rootItems, extension.RootItems()...)

				config, err := config.Load()
				if err != nil {
					return nil, nil, err
				}
				for _, rootItem := range config.Root {
					if rootItem.Extension != extension.Alias {
						continue
					}

					rootItems = append(rootItems, rootItem)
				}

				return exts, rootItems, nil
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
				input := types.CommandInput{
					Command: command.Name,
					Params:  make(map[string]any),
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

						if !param.Required {
							continue
						}

						return fmt.Errorf("%s is a required parameter", param.Name)
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
					case types.ParamTypeNumber:
						value, err := cmd.Flags().GetInt(param.Name)
						if err != nil {
							return err
						}
						input.Params[param.Name] = value
					}
				}

				if !isatty.IsTerminal(os.Stdout.Fd()) {
					output, err := extension.Output(input)

					if err != nil {
						return err
					}

					if _, err := os.Stdout.Write(output); err != nil {
						return err
					}

					return nil
				}

				switch command.Mode {
				case types.CommandModePage:
					runner := tui.NewRunner(extension, input)
					return tui.Draw(runner)
				case types.CommandModeSilent:
					return extension.Run(input)
				case types.CommandModeTTY:
					cmd, err := extension.Cmd(input)
					if err != nil {
						return err
					}
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr

					return cmd.Run()
				default:
					return fmt.Errorf("unknown command mode: %s", command.Mode)
				}

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
