package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
)

func Execute(version string) (err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var config tui.Config
	configPath := path.Join(homeDir, ".config", "sunbeam", "config.yml")
	if _, err := os.Stat(configPath); err == nil {
		bytes, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		if err := yaml.Unmarshal(bytes, &config); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	extensionRoot := path.Join(homeDir, ".local", "share", "sunbeam", "extensions")
	if _, err := os.Stat(extensionRoot); os.IsNotExist(err) {
		if err := os.MkdirAll(extensionRoot, 0755); err != nil {
			return err
		}
	}

	extensions, err := app.LoadExtensions(extensionRoot)
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
			rootList := tui.NewRootList(extensions...)
			model := tui.NewModel(rootList)
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
	rootCmd.AddCommand(NewCmdExtension(extensionRoot, extensions))
	rootCmd.AddCommand(NewCmdServe(extensions))
	rootCmd.AddCommand(NewCmdCheck())
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewCmdRun(&config))
	rootCmd.AddCommand(NewCmdHttp())

	for _, extension := range extensions {
		rootCmd.AddCommand(NewExtensionCommand(extension, &config))
	}

	return rootCmd.Execute()
}

func NewExtensionCommand(extension *app.Extension, config *tui.Config) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:     extension.Name(),
		GroupID: "extension",
		Short:   extension.Description,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rootList := tui.NewRootList(extension)
			model := tui.NewModel(rootList)
			err = tui.Draw(model)
			if err != nil {
				return fmt.Errorf("could not run extension: %w", err)
			}

			return nil
		},
	}

	for _, command := range extension.Commands {
		command := command
		scriptCmd := &cobra.Command{
			Use:   command.Name,
			Short: command.Description,
			RunE: func(cmd *cobra.Command, args []string) (err error) {
				with := make(map[string]app.Arg)
				for _, param := range command.Params {
					if !cmd.Flags().Changed(param.Name) {
						continue
					}

					switch param.Type {
					case "boolean":
						value, err := cmd.Flags().GetBool(param.Name)
						if err != nil {
							return err
						}
						with[param.Name] = app.Arg{Value: value}
					default:
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}
						with[param.Name] = app.Arg{Value: value}
					}

				}
				runner := tui.NewCommandRunner(
					extension,
					command,
					with,
				)

				model := tui.NewModel(runner)

				err = tui.Draw(model)
				if err != nil {
					return fmt.Errorf("could not run script: %w", err)
				}
				return nil
			},
		}

		for _, param := range command.Params {
			switch param.Type {
			case "boolean":
				if param.Default != nil {
					defaultValue := param.Default.(bool)
					scriptCmd.Flags().Bool(param.Name, defaultValue, param.Description)
				} else {
					scriptCmd.Flags().Bool(param.Name, false, param.Description)
				}
			default:
				if param.Default != nil {
					defaultValue := param.Default.(string)
					scriptCmd.Flags().String(param.Name, defaultValue, param.Description)
				} else {
					scriptCmd.Flags().String(param.Name, "", param.Description)
				}
			}

			if param.Type == "file" {
				scriptCmd.MarkFlagFilename(param.Name)
			}

			if param.Type == "directory" {
				scriptCmd.MarkFlagDirname(param.Name)
			}
		}

		extensionCmd.AddCommand(scriptCmd)
	}

	return extensionCmd
}
