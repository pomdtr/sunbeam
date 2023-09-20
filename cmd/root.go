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
			Height: 0,
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

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	cachePath := filepath.Join(cacheDir, "sunbeam", "extensions.json")

	extensions, err := tui.LoadExtensions(config, cachePath)
	if err != nil {
		return nil, err
	}

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		Version:      fmt.Sprintf("%s (%s)", Version, Date),
		SilenceUsage: true,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isatty.IsTerminal(os.Stdout.Fd()) {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(extensions); err != nil {
					return err
				}
			}

			return tui.Draw(tui.NewRootPage(extensions), config.Window)
		},
	}

	rootCmd.AddGroup(
		&cobra.Group{ID: coreGroupID, Title: "Core Commands"},
	)

	rootCmd.AddCommand(NewCmdUpdate(config, cachePath))
	rootCmd.AddCommand(NewCmdRun(extensions, config.Window))
	rootCmd.AddCommand(NewCmdServe(extensions))
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

	if len(extensions.List()) == 0 {
		return rootCmd, nil
	}

	rootCmd.AddGroup(
		&cobra.Group{ID: extensionGroupID, Title: "Extension Commands"},
	)
	for _, alias := range extensions.List() {
		cmd, err := NewCustomCmd(extensions, alias, config.Window)
		if err != nil {
			return nil, err
		}
		rootCmd.AddCommand(cmd)
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

func NewCustomCmd(extensions tui.Extensions, alias string, options tui.WindowOptions) (*cobra.Command, error) {
	extension, ok := extensions[alias]
	if !ok {
		return nil, fmt.Errorf("extension %s does not exist", alias)
	}

	cmd := &cobra.Command{
		Use:          alias,
		Short:        extension.Manifest.Title,
		Long:         extension.Manifest.Description,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isatty.IsTerminal(os.Stdout.Fd()) {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(extension); err != nil {
					return err
				}
			}

			return tui.Draw(tui.NewRootPage(extensions, alias), options)
		},
		GroupID: extensionGroupID,
	}

	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	for name, command := range extension.Manifest.Commands {
		name := name
		command := command
		subcmd := &cobra.Command{
			Use:   name,
			Short: command.Title,
			Long:  command.Description,
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				args := make(map[string]any)
				for name, arg := range command.Params {
					switch arg.Type {
					case types.ParamTypeString:
						value, err := cmd.Flags().GetString(name)
						if err != nil {
							return err
						}

						args[name] = value
					case types.ParamTypeBoolean:
						value, err := cmd.Flags().GetBool(name)
						if err != nil {
							return err
						}

						args[name] = value
					default:
						return fmt.Errorf("unsupported argument type: %s", arg.Type)
					}
				}

				if !isatty.IsTerminal(os.Stdout.Fd()) {
					output, err := extension.Run(name, tui.CommandInput{
						Params: args,
					})
					if err != nil {
						return err
					}

					os.Stdout.Write(output)
					return nil
				}

				return tui.Draw(tui.NewCommand(extensions, types.CommandRef{
					Extension: alias,
					Name:      name,
					Params:    args,
				}), options)
			},
		}

		for name, arg := range command.Params {
			switch arg.Type {
			case types.ParamTypeString:
				subcmd.Flags().String(name, "", arg.Description)
			case types.ParamTypeBoolean:
				subcmd.Flags().Bool(name, false, arg.Description)
			default:
				return nil, fmt.Errorf("unsupported argument type: %s", arg.Type)
			}

			if !arg.Optional {
				subcmd.MarkFlagRequired(name)
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
