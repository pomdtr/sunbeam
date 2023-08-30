package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type ExitCodeError struct {
	ExitCode int
}

func (e ExitCodeError) Error() string {
	return fmt.Sprintf("exit code %d", e.ExitCode)
}

// TODO: cache dependencies
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run <origin> [args...]",
		Short:   "read a sunbeam command",
		GroupID: coreGroupID,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// build cobra command dynamically
			origin, err := originToUrl(args[0])
			if err != nil {
				return fmt.Errorf("could not parse origin: %w", err)
			}

			remote, err := GetRemote(origin)
			if err != nil {
				return fmt.Errorf("could not get remote: %w", err)
			}

			version, err := remote.GetLatestVersion()
			if err != nil {
				return fmt.Errorf("could not get latest version: %w", err)
			}

			var command Extension
			switch remote := remote.(type) {
			case LocalRemote:
				manifest, err := LoadManifest(origin.Path)
				if err != nil {
					return fmt.Errorf("could not load manifest: %w", err)
				}

				command = Extension{
					Manifest: manifest,
					Metadata: Metadata{
						Version: version,
						RootDir: origin.Path,
						Origin:  origin.Path,
					},
				}
			default:
				tempdir, err := os.MkdirTemp("", "sunbeam")
				if err != nil {
					return fmt.Errorf("could not create tempdir: %w", err)
				}
				defer os.RemoveAll(tempdir)
				if _, err := remote.Download(tempdir, ""); err != nil {
					return fmt.Errorf("could not download remote: %w", err)
				}
			}

			tempCmd, err := NewCustomCmd("run", command)
			if err != nil {
				return err
			}

			tempCmd.SetArgs(args)
			return tempCmd.Execute()
		},
	}

	return cmd
}
