package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		Version:      fmt.Sprintf("%s (%s)", Version, Date),
		SilenceUsage: true,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
	}

	rootCmd.AddGroup(
		&cobra.Group{ID: coreGroupID, Title: "Core Commands"},
	)

	configDir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	extensions, err := internal.LoadExtensions(filepath.Join(configDir, "extensions.json"))
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(NewCmdRun(extensions))
	rootCmd.AddCommand(NewExtensionCmd(extensions))
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
	for _, name := range extensions.List() {
		cmd, err := NewCustomCmd(extensions, name)
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

func NewCustomCmd(extensions internal.Extensions, extensionName string) (*cobra.Command, error) {
	extension, err := extensions.Get(extensionName)
	if err != nil {
		return nil, err
	}

	cmd := &cobra.Command{
		Use:          extensionName,
		Short:        extension.Manifest.Title,
		Long:         extension.Manifest.Description,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		GroupID:      extensionGroupID,
	}

	rootCmds := make([]pkg.Command, 0)
	for _, command := range extension.Manifest.Commands {
		if len(command.Params) == 0 {
			rootCmds = append(rootCmds, command)
		}
	}

	if len(rootCmds) == 1 {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return Run(extensions, extensionName, rootCmds[0].Name, nil)
		}
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

				return Run(extensions, extensionName, command.Name, params)
			},
		}

		for _, param := range command.Params {
			switch param.Type {
			case pkg.ParamTypeString:
				var defaultValue string
				if param.Default != nil {
					d, ok := param.Default.(string)
					if !ok {
						return nil, fmt.Errorf("invalid default value for %s: %v", param.Name, param.Default)
					}
					defaultValue = d
				}
				subcmd.Flags().String(param.Name, defaultValue, param.Description)
			case pkg.ParamTypeBoolean:
				var defaultValue bool
				if param.Default != nil {
					d, ok := param.Default.(bool)
					if !ok {
						return nil, fmt.Errorf("invalid default value for %s: %v", param.Name, param.Default)
					}
					defaultValue = d
				}
				subcmd.Flags().Bool(param.Name, defaultValue, param.Description)
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

func Run(extensions internal.Extensions, extensionName string, commandName string, params map[string]any) error {
	extension, err := extensions.Get(extensionName)
	if err != nil {
		return err
	}

	command, ok := extension.Command(commandName)
	if !ok {
		return fmt.Errorf("command %s not found", commandName)
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		output, err := extension.Run(command.Name, pkg.CommandInput{
			Params: params,
		})
		if err != nil {
			return err
		}

		os.Stdout.Write(output)
		return nil
	}

	page, err := internal.CommandToPage(extensions, extensionName, command.Name, params)
	if err != nil {
		return err
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return fmt.Errorf("sunbeam can only be run in a terminal")
	}

	if err := internal.Draw(page, internal.SunbeamOptions{}); err == nil {
		return nil
	} else if errors.Is(err, internal.ErrInterrupted) && isatty.IsTerminal(os.Stdout.Fd()) {
		return nil
	} else {
		return err
	}
}
