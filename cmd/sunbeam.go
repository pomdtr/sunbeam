package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	cobracompletefig "github.com/withfig/autocomplete-tools/integrations/cobra"
)

func parseConfig(configRoot string) (*tui.Config, error) {
	viper.AddConfigPath(configRoot)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.SetEnvPrefix("sunbeam")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	var config tui.Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, err
}

func Execute(version string) (err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	config, err := parseConfig(path.Join(homeDir, ".config", "sunbeam"))
	if err != nil {
		return err
	}

	extensionRoot := path.Join(homeDir, ".local", "share", "sunbeam", "extensions")
	if _, err := os.Stat(extensionRoot); os.IsNotExist(err) {
		if err := os.MkdirAll(extensionRoot, 0755); err != nil {
			return err
		}
	}

	api := app.Api{}
	err = api.LoadExtensions(extensionRoot)
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
			model := tui.NewModel(config, api.Extensions...)
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
	rootCmd.AddCommand(cobracompletefig.CreateCompletionSpecCommand())
	rootCmd.AddCommand(NewCmdDocs())
	rootCmd.AddCommand(NewCmdExtension(api, config))
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewCmdRun(config))

	if os.Getenv("DISABLE_EXTENSIONS") == "" {
		// Extension Commands
		for _, extension := range api.Extensions {
			rootCmd.AddCommand(NewExtensionCommand(extension, config))
		}
	}

	return rootCmd.Execute()
}

func NewExtensionCommand(extension app.Extension, config *tui.Config) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:     extension.Name,
		GroupID: "extension",
		Short:   extension.Description,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			root := tui.NewModel(config, extension)
			err = tui.Draw(root)
			if err != nil {
				return fmt.Errorf("could not run extension: %w", err)
			}

			return nil
		},
	}

	for key, script := range extension.Commands {
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

				model := tui.NewModel(config, extension)
				runner := tui.NewScriptRunner(extension, script, with)
				model.SetRoot(runner)

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
