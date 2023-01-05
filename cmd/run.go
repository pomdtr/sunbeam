package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sunbeamlauncher/sunbeam/app"
	"github.com/sunbeamlauncher/sunbeam/tui"
)

func NewCmdRun() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run <extension> [script] [params]",
		Short: "Run a script",
		Args:  cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
	}

	// Extensions
	for _, extension := range app.Sunbeam.Extensions {
		extensionCmd := NewExtensionCommand(extension)
		runCmd.AddCommand(extensionCmd)
	}

	return runCmd
}

func NewExtensionCommand(extension app.Extension) *cobra.Command {
	extensionCmd := &cobra.Command{
		Use:   extension.Name,
		Short: extension.Description,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			list := tui.NewRootList(extension.RootItems...)
			root := tui.NewModel(list)
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
				model := tui.NewModel(runner)
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
