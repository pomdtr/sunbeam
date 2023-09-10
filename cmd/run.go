package cmd

import (
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/spf13/cobra"
)

func NewCmdRun(extensions internal.Extensions) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <origin> [args...]",
		Short:              "Run an extension without installing it",
		Args:               cobra.MinimumNArgs(1),
		GroupID:            coreGroupID,
		DisableFlagParsing: true,
		SilenceErrors:      true,
		RunE: func(_ *cobra.Command, args []string) error {
			origin, err := parseOrigin(args[0])
			if err != nil {
				return err
			}

			manifest, err := internal.LoadManifest(origin)
			if err != nil {
				return err
			}

			extensionName := filepath.Base(origin.Path)
			extensions.Add(extensionName, internal.Extension{
				Origin:   origin.String(),
				Manifest: manifest,
			})

			cmd, err := NewCustomCmd(extensions, extensionName)
			if err != nil {
				return err
			}

			cmd.SetArgs(args[1:])
			return cmd.Execute()
		},
	}

	return cmd
}
