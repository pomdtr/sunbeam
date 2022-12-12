package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
)

func ParseConfig() tui.Config {
	viper.AddConfigPath(app.Sunbeam.ConfigRoot)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("sunbeam")
	viper.AutomaticEnv()
	viper.ReadInConfig()

	var config tui.Config

	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		log.Printf("No config file found, using default config")
		return config
	}

	bytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("could not read config file: %v", err)
	}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Fatalf("could not parse config file: %v", err)
	}
	return config
}

func Execute() (err error) {
	config := ParseConfig()
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:     "sunbeam",
		Short:   "Command Line Launcher",
		Version: app.Version,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			rootItems := make([]app.RootItem, 0)
			for _, extension := range app.Sunbeam.Extensions {
				rootItems = append(rootItems, extension.RootItems...)
			}

			for _, rootItem := range config.RootItems {
				rootItem.Subtitle = "User"
				rootItems = append(rootItems, rootItem)
			}

			rootList := tui.RootList(rootItems...)
			return tui.Draw(rootList, config)
		},
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    "core",
		Title: "Core Commands",
	}, &cobra.Group{
		ID:    "extensions",
		Title: "Extension Commands",
	})

	// Core Commands
	rootCmd.AddCommand(NewCmdExtension(config))
	rootCmd.AddCommand(NewCmdQuery())

	// Extensions
	for _, extension := range app.Sunbeam.Extensions {
		cmd := NewExtensionCommand(extension, config)
		cmd.GroupID = "extensions"
		rootCmd.AddCommand(cmd)
	}

	return rootCmd.Execute()
}

func NewExtensionCommand(extension app.Extension, config tui.Config) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:   extension.Name,
		Short: extension.Description,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			root := tui.RootList(extension.RootItems...)
			err = tui.Draw(root, config)
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
				with := make(map[string]any, 0)

				for _, param := range script.Inputs {
					switch param.Type {
					case "string", "file", "directory":
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}
						with[param.Name] = value
					case "boolean":
						value, err := cmd.Flags().GetBool(param.Name)
						if err != nil {
							return err
						}
						with[param.Name] = value
					}
				}

				container := tui.NewRunContainer(extension, script, with)
				err = tui.Draw(container, config)
				if err != nil {
					return fmt.Errorf("could not run script: %w", err)
				}
				return nil
			},
		}

		for _, param := range script.Inputs {
			switch param.Type {
			case "string", "file", "directory":
				if defaultValue, ok := param.Default.(string); ok {
					scriptCmd.Flags().String(param.Name, defaultValue, param.Description)
				} else {
					scriptCmd.Flags().String(param.Name, "", param.Description)
					scriptCmd.MarkFlagRequired(param.Name)
				}
			case "boolean":
				if defaultValue, ok := param.Default.(bool); ok {
					scriptCmd.Flags().Bool(param.Name, defaultValue, param.Description)
				} else {
					scriptCmd.Flags().Bool(param.Name, false, param.Description)
					scriptCmd.MarkFlagRequired(param.Name)
				}
			}
			if param.Enum != nil {
				choices := make([]string, len(param.Enum))
				for i, choice := range param.Enum {
					choices[i] = fmt.Sprintf("%v", choice)
				}
				scriptCmd.RegisterFlagCompletionFunc(param.Name, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
					return choices, cobra.ShellCompDirectiveNoFileComp
				})
			}
		}

		extensionCmd.AddCommand(scriptCmd)
	}

	return extensionCmd
}
