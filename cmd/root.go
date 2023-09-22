package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/tailscale/hujson"
	"muzzammil.xyz/jsonc"
)

const (
	coreGroupID      = "core"
	extensionGroupID = "extension"
)

var (
	Version = "dev"
	Date    = "unknown"
)

type ExtensionCache map[string]types.Manifest

func LoadConfig() (tui.Config, error) {
	var config tui.Config

	var candidates []string
	if runtime.GOOS == "darwin" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return config, err
		}

		candidates = append(candidates, filepath.Join(homeDir, ".config", "sunbeam", "config.json"))
		candidates = append(candidates, filepath.Join(homeDir, ".config", "sunbeam", "config.jsonc"))
	} else {
		configHome, err := os.UserConfigDir()
		if err != nil {
			return config, err
		}

		candidates = append(candidates, filepath.Join(configHome, "sunbeam", "config.json"))
		candidates = append(candidates, filepath.Join(configHome, "sunbeam", "config.jsonc"))
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}

		bytes, err := os.ReadFile(candidate)
		if err != nil {
			return config, err
		}

		if strings.HasSuffix(candidate, ".jsonc") {
			b, err := hujson.Standardize(bytes)
			if err != nil {
				return config, err
			}
			bytes = b
		}

		if err := jsonc.Unmarshal(bytes, &config); err != nil {
			return config, err
		}

		return config, nil
	}

	return tui.Config{
		Extensions: make(map[string]string),
		Window: tui.WindowOptions{
			Height: 25,
			Margin: 0,
			Border: true,
		},
	}, nil
}

func NewRootCmd() (*cobra.Command, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	config.Window.Height = LookupIntEnv("SUNBEAM_HEIGHT", config.Window.Height)
	config.Window.Margin = LookupIntEnv("SUNBEAM_MARGIN", config.Window.Margin)
	config.Window.Border = LookupBoolEnv("SUNBEAM_BORDER", config.Window.Border)

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		Version:      fmt.Sprintf("%s (%s)", Version, Date),
		SilenceUsage: true,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
	}

	if len(config.Items) > 0 {
		rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
			items := make([]types.RootItem, 0)
			for _, item := range config.Items {
				origin, ok := config.Extensions[item.Extension]
				if !ok {
					continue
				}

				item.Origin = origin
				items = append(items, item)
			}

			list := tui.NewRootList("Sunbeam", items...)
			return tui.Draw(list, config.Window)
		}

	}

	rootCmd.AddGroup(
		&cobra.Group{ID: coreGroupID, Title: "Core Commands"},
	)

	rootCmd.AddCommand(NewCmdRun(config))
	// rootCmd.AddCommand(NewCmdServe(extensions))
	rootCmd.AddCommand(NewValidateCmd())

	docCmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation for sunbeam",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := buildDoc(rootCmd)
			if err != nil {
				return err
			}

			fmt.Println(doc)
			return nil
		},
	}
	rootCmd.AddCommand(docCmd)

	manCmd := &cobra.Command{
		Use:    "generate-man-pages [path]",
		Short:  "Generate Man Pages for sunbeam",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			header := &doc.GenManHeader{
				Title:   "MINE",
				Section: "3",
			}
			err := doc.GenManTree(rootCmd, header, args[0])
			if err != nil {
				return err
			}

			return nil
		},
	}
	rootCmd.AddCommand(manCmd)

	rootCmd.AddGroup(
		&cobra.Group{ID: extensionGroupID, Title: "Extension Commands"},
	)
	for alias, origin := range config.Extensions {
		alias := alias
		origin := origin
		rootCmd.AddCommand(&cobra.Command{
			Use:                alias,
			Short:              origin,
			DisableFlagParsing: true,
			GroupID:            extensionGroupID,
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				extension, err := tui.LoadExtension(origin)
				if err != nil {
					return nil, cobra.ShellCompDirectiveError
				}

				if len(args) == 0 {
					commands := make([]string, 0, len(extension.Manifest.Commands))
					for _, command := range extension.Manifest.Commands {
						commands = append(commands, command.Name)
					}
					return commands, cobra.ShellCompDirectiveNoFileComp
				}

				return []string{}, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) == 0 {
					return tui.Draw(tui.NewExtensionPage(origin), config.Window)
				}

				extension, err := tui.LoadExtension(origin)
				if err != nil {
					return err
				}
				extension.Alias = alias

				extensionCmd, err := NewCustomCmd(extension, config)
				if err != nil {
					return err
				}

				extensionCmd.SilenceErrors = true
				extensionCmd.SetArgs(args)

				return extensionCmd.Execute()
			},
		})
	}

	return rootCmd, nil
}

func buildDoc(command *cobra.Command) (string, error) {
	if command.GroupID == extensionGroupID {
		return "", nil
	}

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

func NewCustomCmd(extension tui.Extension, config tui.Config) (*cobra.Command, error) {
	use := extension.Alias
	if use == "" {
		use = extension.Origin.String()
	}

	cmd := &cobra.Command{
		Use:          use,
		Short:        extension.Manifest.Title,
		Long:         extension.Manifest.Description,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}

	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	for _, command := range extension.Manifest.Commands {
		command := command
		subcmd := &cobra.Command{
			Use:   command.Name,
			Short: command.Title,
			Long:  command.Description,
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
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
					default:
						return fmt.Errorf("unsupported argument type: %s", param.Type)
					}
				}

				if !isatty.IsTerminal(os.Stdout.Fd()) {
					output, err := extension.Run(command.Name, tui.CommandInput{
						Params: params,
					})
					if err != nil {
						return err
					}

					os.Stdout.Write(output)
					return nil
				}

				extensions := tui.Extensions{
					extension.Origin.String(): extension,
				}

				if command.Mode == types.CommandModeSilent {
					_, err := extension.Run(command.Name, tui.CommandInput{
						Params: params,
					})

					return err
				}

				if command.Mode == types.CommandModeAction {
					output, err := extension.Run(command.Name, tui.CommandInput{
						Params: params,
					})
					if err != nil {
						return err
					}

					var action types.Action
					if err := json.Unmarshal(output, &action); err != nil {
						return err
					}

					return tui.RunAction(action)
				}

				return tui.Draw(
					tui.NewCommand(
						extensions,
						types.CommandRef{
							Origin: extension.Origin.String(),
							Name:   command.Name,
							Params: params,
						}),
					config.Window,
				)
			},
		}

		for _, param := range command.Params {
			switch param.Type {
			case types.ParamTypeString:
				subcmd.Flags().String(param.Name, "", param.Description)
			case types.ParamTypeBoolean:
				subcmd.Flags().Bool(param.Name, false, param.Description)
			default:
				return nil, fmt.Errorf("unsupported argument type: %s", param.Type)
			}

			if !param.Optional {
				subcmd.MarkFlagRequired(param.Name)
			}
		}

		cmd.AddCommand(subcmd)
	}

	return cmd, nil
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
