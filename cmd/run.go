package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewCmdRun(keystore *tui.KeyStore, history *tui.History, config *tui.Config) *cobra.Command {
	runCmd := &cobra.Command{
		Use:     "run <extension-root>",
		Short:   "Run an extension from a directory",
		GroupID: "core",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionRoot := args[0]
			if fi, err := os.Stat(extensionRoot); os.IsNotExist(err) {
				return fmt.Errorf("directory %s does not exist", extensionRoot)
			} else if !fi.IsDir() {
				return fmt.Errorf("directory %s is not a sunbeam extension", extensionRoot)
			}
			manifestPath := filepath.Join(extensionRoot, "sunbeam.yml")
			if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
				return fmt.Errorf("directory %s is not a sunbeam extension", extensionRoot)
			}

			extension, err := app.ParseManifest(manifestPath)
			extension.Root = extensionRoot
			if err != nil {
				return fmt.Errorf("failed to parse manifest: %s", err)
			}

			rootList := tui.NewRootList(keystore, history, &extension)
			model := tui.NewModel(rootList)

			return tui.Draw(model)
		},
	}

	return runCmd
}
