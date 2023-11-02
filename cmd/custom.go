package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewCmdCustom(alias string, extension extensions.Extension) (*cobra.Command, error) {
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
				return encoder.Encode(extension)
			}

			var rootItems []types.RootItem
			for _, rootItem := range extension.RootItems() {
				rootItem.Extension = alias
				rootItems = append(rootItems, rootItem)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			for _, rootItem := range cfg.Root {
				if rootItem.Extension != alias {
					continue
				}

				rootItems = append(rootItems, rootItem)
			}

			if len(rootItems) == 0 {
				return cmd.Usage()
			}

			page := tui.NewRootList(extension.Title, func() (extensions.ExtensionMap, []types.RootItem, error) {
				extension, err := LoadExtension(extension.Entrypoint)
				if err != nil {
					return nil, nil, err
				}

				var rootItems []types.RootItem
				for _, rootItem := range extension.RootItems() {
					rootItem.Extension = alias
					rootItems = append(rootItems, rootItem)
				}

				cfg, err := config.Load()
				if err != nil {
					return nil, nil, err
				}
				for _, rootItem := range cfg.Root {
					if rootItem.Extension != alias {
						continue
					}

					rootItems = append(rootItems, rootItem)
				}
				return extensions.ExtensionMap{
					alias: extension,
				}, rootItems, nil
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

				missing := tui.FindMissingParams(command.Params, params)
				if !isatty.IsTerminal(os.Stdout.Fd()) {
					if len(missing) > 0 {
						names := make([]string, len(missing))
						for i, param := range missing {
							names[i] = param.Name
						}
						return fmt.Errorf("missing required params: %s", strings.Join(names, ", "))
					}
					output, err := extension.Output(input)

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
