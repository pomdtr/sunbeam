package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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

	return tui.Config{
		Height: viper.GetInt("height"),
	}
}

func Execute() (err error) {
	config := ParseConfig()
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:     "sunbeam",
		Short:   "Command Line Launcher",
		Version: app.Version,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			manifests := make([]app.Extension, 0)
			for _, manifest := range app.Sunbeam.Extensions {
				manifests = append(manifests, manifest)
			}

			rootList := tui.RootList(manifests...)
			err = tui.Draw(rootList, config)
			if err != nil {
				return err
			}
			return
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
	rootCmd.AddCommand(NewRawInputCommand(config))

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
			var runner tui.Container
			// If there is only one root item, just run it
			if len(extension.RootItems) == 1 {
				item := extension.RootItems[0]
				script, ok := extension.Scripts[item.Script]
				if !ok {
					return fmt.Errorf("script %s not found", item.Script)
				}
				runner = tui.NewRunContainer(extension, script, nil)
			} else {
				runner = tui.RootList(extension)
			}
			err = tui.Draw(runner, config)
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
				// with := make([]any, 0)
				for key := range script.Params {
					value, err := cmd.Flags().GetString(key)
					if err != nil {
						return err
					}
					var forminput map[string]interface{}
					if err := json.Unmarshal([]byte(value), &forminput); err == nil {
						// append(
					} else {
						// switch param.Type {
						// case "string":
						// 	appendkk
						// case "bool":
						// 	with[key] = value == "true"
						// }
					}
				}

				container := tui.NewRunContainer(extension, script, nil)
				err = tui.Draw(container, config)
				if err != nil {
					return fmt.Errorf("could not run script: %w", err)
				}
				return nil
			},
		}

		for key, param := range script.Params {
			flag := NewCustomFlag(param)
			scriptCmd.Flags().Var(flag, key, param.Description)
			if param.Default == nil {
				scriptCmd.MarkFlagRequired(key)
			}
			// if param.Enum != nil {
			// 	choices := make([]string, len(param.Enum))
			// 	for i, choice := range param.Enum {
			// 		choices[i] = fmt.Sprintf("%v", choice)
			// 	}
			// 	scriptCmd.RegisterFlagCompletionFunc(key, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// 		return choices, cobra.ShellCompDirectiveNoFileComp
			// 	})
			// }
		}

		extensionCmd.AddCommand(scriptCmd)
	}

	return extensionCmd
}
