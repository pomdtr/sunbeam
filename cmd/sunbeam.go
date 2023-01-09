package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/sunbeamlauncher/sunbeam/app"
	"github.com/sunbeamlauncher/sunbeam/tui"
)

func Execute(version string) error {
	extensionRoot := path.Join(xdg.DataHome, "sunbeam", "extensions")
	if _, err := os.Stat(extensionRoot); os.IsNotExist(err) {
		if err := os.MkdirAll(extensionRoot, 0755); err != nil {
			return err
		}
	}

	api := app.Api{}
	err := api.LoadExtensions(extensionRoot)
	if err != nil {
		return err
	}

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		SilenceUsage: true,
		Version:      version,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rootItems := make([]app.RootItem, 0)
			for _, extension := range api.Extensions {
				rootItems = append(rootItems, extension.RootItems...)
			}

			for _, rootItem := range tui.Config.RootItems {
				rootItem.Subtitle = "User"
				rootItems = append(rootItems, rootItem)
			}

			rootList := tui.NewRootList(rootItems...)
			model := tui.NewModel(rootList, api.Extensions...)
			return tui.Draw(model)
		},
	}

	rootCmd.AddGroup(&cobra.Group{
		Title: "Core Commands",
		ID:    "core",
	}, &cobra.Group{
		Title: "Extension Commands",
		ID:    "extension",
	})

	// Core Commands
	rootCmd.AddCommand(NewCmdExtension(api))
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(NewCmdClipboard())
	rootCmd.AddCommand(NewCmdOpen())

	rootCmd.AddCommand(func() *cobra.Command {
		return &cobra.Command{
			Use:    "generate-docs",
			Args:   cobra.ExactArgs(1),
			Hidden: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				target := args[0]
				if _, err := os.Stat(target); os.IsNotExist(err) {
					if err := os.MkdirAll(target, 0755); err != nil {
						return err
					}
				}
				return doc.GenMarkdownTreeCustom(
					rootCmd,
					target,
					func(s string) string {
						basename := path.Base(s)
						stem := strings.TrimSuffix(basename, ".md")
						title := strings.ReplaceAll(stem, "_", " ")
						return fmt.Sprintf("---\ntitle: %s\nhide_title: true\n---\n\n", title)
					},
					func(s string) string { return fmt.Sprintf("./%s", s) },
				)
			},
		}
	}())

	if os.Getenv("DISABLE_EXTENSIONS") == "" {
		// Extension Commands
		for _, extension := range api.Extensions {
			rootCmd.AddCommand(NewExtensionCommand(extension, api.Extensions))
		}
	}

	return rootCmd.Execute()
}

func NewExtensionCommand(extension app.Extension, extensions []app.Extension) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:     extension.Name,
		GroupID: "extension",
		Short:   extension.Description,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			list := tui.NewRootList(extension.RootItems...)
			root := tui.NewModel(list, extensions...)
			err = tui.Draw(root)
			if err != nil {
				return fmt.Errorf("could not run extension: %w", err)
			}

			return nil
		},
	}

	for key, script := range extension.Scripts {
		script := script
		scriptCmd := &cobra.Command{
			Use:   key,
			Short: script.Description,
			RunE: func(cmd *cobra.Command, args []string) (err error) {
				with := make(map[string]app.ScriptInputWithValue)
				for _, param := range script.Inputs {
					if !cmd.Flags().Changed(param.Name) {
						continue
					}
					switch param.Type {
					case "checkbox":
						value, err := cmd.Flags().GetBool(param.Name)
						if err != nil {
							return err
						}
						with[param.Name] = app.ScriptInputWithValue{Value: value}
					default:
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}
						with[param.Name] = app.ScriptInputWithValue{Value: value}
					}

				}

				runner := tui.NewScriptRunner(extension, script, with)
				model := tui.NewModel(runner, extensions...)
				err = tui.Draw(model)
				if err != nil {
					return fmt.Errorf("could not run script: %w", err)
				}
				return nil
			},
		}

		for _, param := range script.Inputs {
			switch param.Type {
			case "checkbox":
				if defaultValue, ok := param.Default.Value.(bool); ok {
					scriptCmd.Flags().Bool(param.Name, defaultValue, param.Title)
				} else {
					scriptCmd.Flags().Bool(param.Name, false, param.Title)
				}
			default:
				if defaultValue, ok := param.Default.Value.(string); ok {
					scriptCmd.Flags().String(param.Name, defaultValue, param.Title)
				} else {
					scriptCmd.Flags().String(param.Name, "", param.Title)
				}
			}
		}

		extensionCmd.AddCommand(scriptCmd)
	}

	return extensionCmd
}
