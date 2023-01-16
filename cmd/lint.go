package cmd

import (
	"fmt"
	"os"

	"github.com/pomdtr/sunbeam/app"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewCmdLint() *cobra.Command {
	return &cobra.Command{
		Use:     "lint",
		Short:   "Lint a sunbeam extension manifest",
		GroupID: "core",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestBytes, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var m any
			err = yaml.Unmarshal(manifestBytes, &m)
			if err != nil {
				return fmt.Errorf("failed to unmarshal manifest: %w", err)
			}

			err = app.ManifestSchema.Validate(m)
			if err != nil {
				return fmt.Errorf("%#v", err)
			}

			fmt.Println("Manifest is valid")
			return nil
		},
	}
}
