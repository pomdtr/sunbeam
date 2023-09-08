package cmd

import (
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <origin> [args...]",
		Short:              "Run an extension without installing it",
		Args:               cobra.MinimumNArgs(1),
		GroupID:            coreGroupID,
		DisableFlagParsing: true,
		RunE: func(_ *cobra.Command, args []string) error {
			origin, err := parseOrigin(args[0])
			if err != nil {
				return err
			}

			manifest, err := LoadManifest(origin)
			if err != nil {
				return err
			}

			cmd, err := NewCustomCmd(args[0], Extension{
				Origin:   origin.String(),
				Manifest: manifest,
			})
			if err != nil {
				return err
			}

			cmd.SetArgs(args[1:])
			return cmd.Execute()
		},
	}

	return cmd
}
