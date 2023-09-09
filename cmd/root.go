package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	dataDir string
)

func init() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	dataDir = filepath.Join(homedir, ".local", "share", "sunbeam")
}

func Execute() error {
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

	rootCmd.AddCommand(NewCmdRun())
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

	extensions, err := LoadExtensions()
	if err != nil {
		return err
	}
	rootCmd.AddCommand(NewExtensionCmd(extensions))
	rootCmd.AddCommand(NewCmdServe(extensions))

	if len(extensions) == 0 {
		return rootCmd.Execute()
	}

	rootCmd.AddGroup(
		&cobra.Group{ID: extensionGroupID, Title: "Extension Commands"},
	)
	for name, extension := range extensions {
		cmd, err := NewCustomCmd(name, extension)
		if err != nil {
			return err
		}
		rootCmd.AddCommand(cmd)
	}

	return rootCmd.Execute()
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

func NewCustomCmd(name string, extension Extension) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:          name,
		Short:        extension.Manifest.Title,
		Long:         extension.Manifest.Description,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		GroupID:      extensionGroupID,
	}

	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	for _, command := range extension.Manifest.Commands {
		command := command
		subcmd := &cobra.Command{
			Use:    command.Name,
			Short:  command.Title,
			Long:   command.Description,
			Hidden: command.Hidden,
			Args:   cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				params := make(map[string]any)
				for _, param := range command.Params {
					switch param.Type {
					case ParamTypeString:
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}

						params[param.Name] = value
					case ParamTypeBoolean:
						value, err := cmd.Flags().GetBool(param.Name)
						if err != nil {
							return err
						}

						params[param.Name] = value
					default:
						return fmt.Errorf("unsupported argument type: %s", param.Type)
					}
				}

				input := CommandInput{
					Params: params,
				}

				output, err := extension.Run(command.Name, input)
				if err != nil {
					return err
				}

				os.Stdout.Write(output)
				return nil
			},
		}

		for _, param := range command.Params {
			switch param.Type {
			case ParamTypeString:
				var defaultValue string
				if param.Default != nil {
					d, ok := param.Default.(string)
					if !ok {
						return nil, fmt.Errorf("invalid default value for %s: %v", param.Name, param.Default)
					}
					defaultValue = d
				}
				subcmd.Flags().String(param.Name, defaultValue, param.Description)
			case ParamTypeBoolean:
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
