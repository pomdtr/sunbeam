package cmd

import (
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal"
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

			manifest, err := internal.LoadManifest(origin)
			if err != nil {
				return err
			}

			alias := filepath.Base(origin.Path)
			if alias == "/" {
				alias = origin.Hostname()
			}

			extension := internal.Extension{
				Origin:   origin.String(),
				Manifest: manifest,
			}

			scriptCmd, err := NewCustomCmd(alias, extension)
			if err != nil {
				return err
			}
			scriptCmd.SilenceErrors = true

			scriptCmd.SetArgs(args[1:])
			return scriptCmd.Execute()
		},
	}

	return cmd
}
