package cmd

import (
	"fmt"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	runCmd := &cobra.Command{
		Use:     "run <extension> [script] [params]",
		Short:   "Run a script",
		GroupID: "core",
		Args:    cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
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
			list := tui.RootList(extension.RootItems...)
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
				with := make(map[string]app.ScriptInput)
				for _, param := range script.Params {
					if !cmd.Flags().Changed(param.Name) {
						continue
					}
					switch param.Input.Type {
					case "checkbox":
						value, err := cmd.Flags().GetBool(param.Name)
						if err != nil {
							return err
						}
						with[param.Name] = app.ScriptInput{Value: value}
					default:
						value, err := cmd.Flags().GetString(param.Name)
						if err != nil {
							return err
						}
						with[param.Name] = app.ScriptInput{Value: value}
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

		for _, param := range script.Params {
			switch param.Input.Type {
			case "checkbox":
				if defaultValue, ok := param.Input.DefaultValue.(bool); ok {
					scriptCmd.Flags().Bool(param.Name, defaultValue, param.Input.Title)
				} else {
					scriptCmd.Flags().Bool(param.Name, false, param.Input.Title)
				}
			default:
				if defaultValue, ok := param.Input.DefaultValue.(string); ok {
					scriptCmd.Flags().String(param.Name, defaultValue, param.Input.Title)
				} else {
					scriptCmd.Flags().String(param.Name, "", param.Input.Title)
				}
			}
		}

		extensionCmd.AddCommand(scriptCmd)
	}

	return extensionCmd
}
