package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/cli/browser"
	"github.com/google/shlex"
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

func getConfigPath() (string, error) {
	var candidates []string
	if runtime.GOOS == "darwin" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		candidates = append(candidates, filepath.Join(homeDir, ".config", "sunbeam", "config.json"))
		candidates = append(candidates, filepath.Join(homeDir, ".config", "sunbeam", "config.jsonc"))
	} else {
		configHome, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}

		candidates = append(candidates, filepath.Join(configHome, "sunbeam", "config.json"))
		candidates = append(candidates, filepath.Join(configHome, "sunbeam", "config.jsonc"))
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}

		return candidate, nil
	}

	return candidates[0], nil
}

func LoadConfig(configPath string) (tui.Config, error) {
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return tui.Config{}, err
	}

	if strings.HasSuffix(configPath, ".jsonc") {
		b, err := hujson.Standardize(bytes)
		if err != nil {
			return tui.Config{}, err
		}
		bytes = b
	}

	var config tui.Config
	if err := jsonc.Unmarshal(bytes, &config); err != nil {
		return tui.Config{}, err
	}

	return config, nil
}

func NewRootCmd() (*cobra.Command, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		config = tui.Config{}
	}

	config.Window.Height = LookupIntEnv("SUNBEAM_HEIGHT", config.Window.Height)

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		Version:      fmt.Sprintf("%s (%s)", Version, Date),
		SilenceUsage: true,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			items := make([]types.ListItem, 0)
			for title, command := range config.Root {
				ref, err := ExtractCommand(command, config.Aliases)
				if err != nil {
					continue
				}

				items = append(items, types.ListItem{
					Title:       title,
					Id:          title,
					Accessories: []string{command},
					Actions: []types.Action{
						{
							Title: "Run Command",
							OnAction: types.Command{
								Type:    types.CommandTypeRun,
								Origin:  ref.Origin,
								Command: ref.Command,
								Params:  ref.Params,
							},
						},
						{
							Title: "Copy Command",
							Key:   "c",
							OnAction: types.Command{
								Type: types.CommandTypeCopy,
								Text: command,
								Exit: true,
							},
						},
					},
				})
			}

			return tui.Draw(tui.NewRootList(tui.NewExtensions(config.Aliases), items...), config.Window)
		},
	}

	rootCmd.AddGroup(
		&cobra.Group{ID: coreGroupID, Title: "Core Commands"},
	)

	rootCmd.AddCommand(NewCmdConfig(configPath))
	rootCmd.AddCommand(NewCmdRun(config))
	rootCmd.AddCommand(NewCmdServe(config))
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

	for alias, origin := range config.Aliases {
		alias := alias
		origin := origin

		rootCmd.AddCommand(&cobra.Command{
			Use:                alias,
			Short:              origin,
			DisableFlagParsing: true,
			GroupID:            extensionGroupID,
			RunE: func(cmd *cobra.Command, args []string) error {
				origin, err := resolveSpecifiers(origin)
				if err != nil {
					return err
				}

				command, err := NewCustomCommand(origin, config)
				if err != nil {
					return err
				}

				command.Use = alias
				command.SetArgs(args)
				return command.Execute()
			},
		})
	}

	return rootCmd, nil
}

func NewCustomCommand(origin string, config tui.Config) (*cobra.Command, error) {
	extensions := tui.NewExtensions(config.Aliases)

	manifest, err := extensions.Get(origin)
	if err != nil {
		return nil, err
	}

	rootCmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Draw(tui.NewRunner(extensions, tui.CommandRef{
				Origin: origin,
			}), config.Window)
		},
		SilenceErrors: true,
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	for _, subcommand := range manifest.Commands {
		subcommand := subcommand
		subcmd := &cobra.Command{
			Use: subcommand.Name,
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
					}
				}

				extension, err := extensions.Get(origin)
				if err != nil {
					return err
				}

				if subcommand.Mode == types.CommandModeView {
					return tui.Draw(tui.NewRunner(extensions, tui.CommandRef{
						Origin:  origin,
						Command: subcommand.Name,
						Params:  params,
					}), config.Window)
				}

				out, err := extension.Run(tui.CommandInput{
					Command: subcommand.Name,
					Params:  params,
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
					return browser.OpenURL(command.Url)
				default:
					return nil
				}
			},
		}

		if subcommand.Hidden {
			subcmd.Hidden = true
		}

		for _, param := range subcommand.Params {
			switch param.Type {
			case types.ParamTypeString:
				subcmd.Flags().String(param.Name, "", param.Description)
			case types.ParamTypeBoolean:
				subcmd.Flags().Bool(param.Name, false, param.Description)
			}

			if !param.Optional {
				subcmd.MarkFlagRequired(param.Name)
			}
		}

		rootCmd.AddCommand(subcmd)
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

func ExtractCommand(exec string, aliases map[string]string) (tui.CommandRef, error) {
	var ref tui.CommandRef
	args, err := shlex.Split(exec)
	if err != nil {
		return ref, err
	}

	if len(args) == 0 {
		return ref, fmt.Errorf("no command specified")
	}

	if args[0] != "sunbeam" {
		return ref, fmt.Errorf("only sunbeam commands are supported")
	}

	if len(args) == 1 {
		return ref, fmt.Errorf("no command specified")
	}

	args = args[1:]

	var origin string
	if args[0] == "run" {
		args = args[1:]
		if len(args) == 0 {
			return ref, fmt.Errorf("no origin specified")
		}

		origin = args[0]
		args = args[1:]
	} else {
		o, ok := aliases[args[0]]
		if !ok {
			return ref, fmt.Errorf("alias %s not found", args[0])
		}

		origin = o
		args = args[1:]
	}

	ref.Origin, err = resolveSpecifiers(origin)
	if err != nil {
		return ref, err
	}

	if len(args) == 0 {
		return ref, nil
	}

	ref.Command = args[0]
	args = args[1:]

	if len(args) == 0 {
		return ref, nil
	}

	ref.Params = make(map[string]any)

	for len(args) > 0 {
		if !strings.HasPrefix(args[0], "--") {
			return ref, fmt.Errorf("invalid argument: %s", args[0])
		}

		arg := strings.TrimPrefix(args[0], "--")

		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			ref.Params[parts[0]] = parts[1]
			args = args[1:]
			continue
		}

		if len(args) == 1 {
			ref.Params[arg] = true
			args = args[1:]
			continue
		}

		if strings.HasPrefix(args[1], "--") {
			ref.Params[arg] = true
			args = args[1:]
			continue
		}

		ref.Params[arg] = args[1]
		args = args[2:]
	}

	return ref, nil
}
