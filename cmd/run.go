package cmd

import (
	"fmt"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewCmdRun(config tui.Config) *cobra.Command {
	runCmd := &cobra.Command{
		Use:     "run <extension> [script] [params]",
		Short:   "Run a script",
		GroupID: "core",
		Args:    cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
	}

	// Extensions
	for _, extension := range app.Sunbeam.Extensions {
		extensionCmd := NewExtensionCommand(extension, config)
		runCmd.AddCommand(extensionCmd)
	}

	return runCmd
}

func NewExtensionCommand(extension app.Extension, config tui.Config) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:   extension.Name,
		Short: extension.Description,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			list := tui.RootList(extension.RootItems...)
			root := tui.NewModel(list, config)
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
				with := make(app.ScriptInputs)
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
						with[param.Name] = app.ScriptParam{Value: value}
					default:
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}
						with[param.Name] = app.ScriptParam{Value: value}
					}

				}

				runner := tui.NewScriptRunner(extension, script, with)
				model := tui.NewModel(runner, config)
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
				if defaultValue, ok := param.Default.(bool); ok {
					scriptCmd.Flags().Bool(param.Name, defaultValue, param.Title)
				} else {
					scriptCmd.Flags().Bool(param.Name, false, param.Title)
				}
			default:
				if defaultValue, ok := param.Default.(string); ok {
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
