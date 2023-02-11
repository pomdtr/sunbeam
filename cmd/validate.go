package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pomdtr/sunbeam/app"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewCmdValidate() *cobra.Command {
	validateCmd := cobra.Command{
		Use:     "validate",
		GroupID: "core",
		Args:    cobra.ExactArgs(1),
	}

	validateCmd.AddCommand(&cobra.Command{
		Use: "manifest <manifest-path>",
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := args[0]
			if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
				return fmt.Errorf("file %s does not exist", manifestPath)
			}

			manifestBytes, err := os.ReadFile(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var m any
			if err = yaml.Unmarshal(manifestBytes, &m); err != nil {
				return fmt.Errorf("failed to unmarshal manifest: %w", err)
			}

			if err = app.ExtensionSchema.Validate(m); err != nil {
				return fmt.Errorf("%#v", err)
			}

			var extension app.Extension
			if err := yaml.Unmarshal(manifestBytes, &extension); err != nil {
				return fmt.Errorf("failed to unmarshal manifest: %w", err)
			}

			for _, rootItem := range extension.RootItems {
				if _, ok := extension.GetCommand(rootItem.Command); !ok {
					return fmt.Errorf("root item '%s' references unknown command '%s'", rootItem.Title, rootItem.Command)
				}
			}

			fmt.Println("Extension is valid")
			return nil
		},
	})

	validateCmd.AddCommand(func() *cobra.Command {
		command := cobra.Command{
			Use: "page",
			RunE: func(cmd *cobra.Command, args []string) error {
				bytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				var m any
				if err = yaml.Unmarshal(bytes, &m); err != nil {
					return fmt.Errorf("failed to unmarshal manifest: %w", err)
				}

				if err = app.PageSchema.Validate(m); err != nil {
					return fmt.Errorf("%#v", err)
				}

				var page app.Page
				if err := json.Unmarshal(bytes, &page); err != nil {
					return err
				}

				if cmd.Flags().Changed("manifest") {
					manifestPath, err := cmd.Flags().GetString("manifest")
					if err != nil {
						return err
					}

					if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
						return err
					}

					manifest, err := os.ReadFile(manifestPath)
					if err != nil {
						return err
					}

					var extension app.Extension
					if err := yaml.Unmarshal(manifest, &extension); err != nil {
						return err
					}

					switch page.Type {
					case "detail":
						for _, action := range page.Actions {
							if action.Type != "run-command" {
								continue
							}

							command, ok := extension.GetCommand(action.Command)
							if !ok {
								return fmt.Errorf("command %s is not defined in manifest", action.Command)
							}

							commandInputMap := make(map[string]struct{})
							for _, param := range command.Params {
								commandInputMap[param.Name] = struct{}{}
							}

							for key := range action.With {
								if _, ok := commandInputMap[key]; !ok {
									return fmt.Errorf("input %s is not defined for command %s", key, action.Command)
								}
							}
						}
					case "list":
						for _, listitem := range page.List.Items {
							for _, action := range listitem.Actions {
								if action.Type != "run-command" {
									continue
								}

								command, ok := extension.GetCommand(action.Command)
								if !ok {
									return fmt.Errorf("command %s is not defined in manifest", action.Command)
								}

								commandInputMap := make(map[string]struct{})
								for _, param := range command.Params {
									commandInputMap[param.Name] = struct{}{}
								}

								for key := range action.With {
									if _, ok := commandInputMap[key]; !ok {
										return fmt.Errorf("input %s is not defined for command %s", key, action.Command)
									}
								}
							}
						}
					}
				}

				fmt.Println("Page is valid")
				return nil
			},
		}

		command.Flags().String("manifest", "", "Path to the manifest file")
		return &command
	}())

	return &validateCmd
}
