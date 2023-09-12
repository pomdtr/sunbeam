package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const (
	coreGroupID      = "core"
	extensionGroupID = "extension"
)

var (
	Version = "dev"
	Date    = "unknown"
)

func getOptions() internal.SunbeamOptions {
	return internal.SunbeamOptions{
		MaxHeight:  LookupIntEnv("SUNBEAM_HEIGHT", 30),
		MaxWidth:   LookupIntEnv("SUNBEAM_WIDTH", 100),
		FullScreen: LookupBoolEnv("SUNBEAM_FULLSCREEN", true),
		Border:     LookupBoolEnv("SUNBEAM_BORDER", true),
		Margin:     LookupIntEnv("SUNBEAM_MARGIN", 1),
		NoColor:    LookupBoolEnv("NO_COLOR", false),
	}
}

func ConfigDir() (string, error) {
	if env, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		return filepath.Join(env, "sunbeam"), nil
	}

	switch runtime.GOOS {
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(homeDir, ".config", "sunbeam"), nil
	default:
		configHome, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(configHome, "sunbeam"), nil
	}
}

func NewRootCmd() (*cobra.Command, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	extensions, err := internal.LoadExtensions(filepath.Join(configDir, "extensions.json"))
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
			return internal.Draw(internal.NewRootPage(extensions.Map()), getOptions())
		},
	}

	rootCmd.AddGroup(
		&cobra.Group{ID: coreGroupID, Title: "Core Commands"},
	)

	rootCmd.AddCommand(NewExtensionCmd(extensions))
	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(NewCmdServe(extensions))
	rootCmd.AddCommand(NewCmdParse())
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
	for alias, extension := range extensions.Map() {
		cmd, err := NewCustomCmd(alias, extension)
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

func NewCustomCmd(alias string, extension internal.Extension) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:          alias,
		Short:        extension.Manifest.Title,
		Long:         extension.Manifest.Description,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			extensions := map[string]internal.Extension{
				alias: extension,
			}
			return internal.Draw(internal.NewRootPage(extensions), getOptions())
		},
		GroupID: extensionGroupID,
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
					switch param.Type {
					case pkg.ParamTypeString:
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}

						params[param.Name] = value
					case pkg.ParamTypeBoolean:
						value, err := cmd.Flags().GetBool(param.Name)
						if err != nil {
							return err
						}

						params[param.Name] = value
					default:
						return fmt.Errorf("unsupported argument type: %s", param.Type)
					}
				}

				if os.Getenv("CI") != "" || !isatty.IsTerminal(os.Stdout.Fd()) {
					output, err := extension.Run(command.Name, pkg.CommandInput{
						Params: params,
					})
					if err != nil {
						return err
					}

					os.Stdout.Write(output)
					return nil
				}

				page, err := internal.CommandToPage(extension, pkg.CommandRef{
					Name:   command.Name,
					Params: params,
				})
				if err != nil {
					return err
				}

				return internal.Draw(page, getOptions())
			},
		}

		for _, param := range command.Params {
			switch param.Type {
			case pkg.ParamTypeString:
				subcmd.Flags().String(param.Name, "", param.Description)
			case pkg.ParamTypeBoolean:
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
