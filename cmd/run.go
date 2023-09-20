package cmd

import (
	"path/filepath"

	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/spf13/cobra"
)

func NewCmdRun(extensions tui.Extensions, options tui.WindowOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run <origin> [args...]",
		Short:              "Run an extension without installing it",
		Args:               cobra.MinimumNArgs(1),
		GroupID:            coreGroupID,
		DisableFlagParsing: true,
		RunE: func(_ *cobra.Command, args []string) error {
			origin, err := tui.ParseOrigin(args[0])
			if err != nil {
				return err
			}

			manifest, err := tui.LoadManifest(origin)
			if err != nil {
				return err
			}

			alias := filepath.Base(origin.Path)
			if alias == "/" {
				alias = origin.Hostname()
			}

			extensions[alias] = tui.Extension{
				Origin:   origin,
				Manifest: manifest,
			}

			scriptCmd, err := NewCustomCmd(extensions, alias, options)
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
