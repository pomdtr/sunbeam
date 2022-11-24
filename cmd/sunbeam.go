package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
)

var globalOptions tui.SunbeamOptions

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sunbeam",
	Short: "Command Line Launcher",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		manifests := make([]app.Extension, 0)
		for _, manifest := range app.Sunbeam.Extensions {
			manifests = append(manifests, manifest)
		}

		rootList := tui.RootList(manifests...)
		err = tui.Draw(rootList, globalOptions)
		if err != nil {
			return err
		}
		return
	},
}

func Execute() (err error) {
	rootCmd.PersistentFlags().IntVarP(&globalOptions.Height, "height", "H", 0, "height of the window")

	rootCmd.AddGroup(&cobra.Group{
		ID:    "core",
		Title: "Core Commands",
	}, &cobra.Group{
		ID:    "extensions",
		Title: "Extension Commands",
	})

	// Core Commands
	rootCmd.AddCommand(NewCmdExtension())
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewCmdVersion())

	// Extensions
	for _, extension := range app.Sunbeam.Extensions {
		cmd := NewExtensionCommand(extension)
		cmd.GroupID = "extensions"
		rootCmd.AddCommand(cmd)
	}

	return rootCmd.Execute()
}

func NewExtensionCommand(extension app.Extension) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:   extension.Name,
		Short: extension.Title,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var runner tui.Container
			// If there is only one root item, just run it
			if len(extension.RootItems) == 1 {
				item := extension.RootItems[0]
				script, ok := extension.Scripts[item.Script]
				if !ok {
					return fmt.Errorf("script %s not found", item.Script)
				}
				runner = tui.NewRunContainer(extension, script, item.With)
			} else {
				runner = tui.RootList(extension)
			}
			err = tui.Draw(runner, globalOptions)
			if err != nil {
				return fmt.Errorf("could not run extension: %w", err)
			}

			return nil
		},
	}

	for key, script := range extension.Scripts {
		script := script
		scriptCmd := &cobra.Command{
			Use: key,
			RunE: func(cmd *cobra.Command, args []string) (err error) {
				with := make(map[string]any)
				for _, input := range script.Params {
					switch input.Type {
					case "checkbox":
						with[input.Name], err = cmd.Flags().GetBool(input.Name)
						if err != nil {
							return err
						}
					default:
						with[input.Name], err = cmd.Flags().GetString(input.Name)
						if err != nil {
							return err
						}
					}
				}

				container := tui.NewRunContainer(extension, script, with)
				err = tui.Draw(container, globalOptions)
				if err != nil {
					return fmt.Errorf("could not run script: %w", err)
				}
				return nil
			},
		}

		for _, input := range script.Params {
			flag := NewCustomFlag(input)
			scriptCmd.Flags().Var(flag, input.Name, input.Title)
			if input.Type == "dropdown" {
				choices := make([]string, len(input.Data))
				for i, choice := range input.Data {
					choices[i] = fmt.Sprintf("%s\t%s", choice.Value, choice.Title)
				}
				scriptCmd.RegisterFlagCompletionFunc(input.Name, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
					return choices, cobra.ShellCompDirectiveNoFileComp
				})

			}
		}

		extensionCmd.AddCommand(scriptCmd)
	}

	return extensionCmd
}
