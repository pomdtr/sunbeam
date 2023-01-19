package cmd

import (
	"fmt"
	"os"

	"github.com/pomdtr/sunbeam/app"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewCmdCheck() *cobra.Command {
	return &cobra.Command{
		Use:     "check <manifest-path>",
		Short:   "Check if an extension manifest is valid",
		GroupID: "core",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := args[0]
			if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "File %s does not exist\n", manifestPath)
				os.Exit(1)
			}

			manifestBytes, err := os.ReadFile(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var m any
			if err = yaml.Unmarshal(manifestBytes, &m); err != nil {
				return fmt.Errorf("failed to unmarshal manifest: %w", err)
			}

			if err = app.ManifestSchema.Validate(m); err != nil {
				return fmt.Errorf("%#v", err)
			}

			var extension app.Extension
			if err := yaml.Unmarshal(manifestBytes, &extension); err != nil {
				return fmt.Errorf("failed to unmarshal manifest: %w", err)
			}

			for _, rootItem := range extension.RootItems {
				if _, ok := extension.Commands[rootItem.Command]; !ok {
					return fmt.Errorf("root item '%s' references unknown command '%s'", rootItem.Title, rootItem.Command)
				}
			}

			fmt.Println("Extension is valid")
			return nil
		},
	}
}
