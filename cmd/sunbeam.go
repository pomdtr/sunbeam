package cmd

import (
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

	var config tui.Config

	err := viper.Unmarshal(&config)
	if err != nil {
		panic(err)
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
			model := tui.NewModel(rootList, config)
			return tui.Draw(model)
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
	rootCmd.AddCommand(NewCmdRun(config))

	return rootCmd.Execute()
}
