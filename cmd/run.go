package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewCmdRun(config *tui.Config) *cobra.Command {
	runCmd := &cobra.Command{
		Use:     "run <extension-root>",
		Short:   "Run an extension from a directory",
		GroupID: "core",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("no extension in current directory")
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		return runCmd
	}

	manifestPath := filepath.Join(cwd, "sunbeam.yml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return runCmd
	}

	extension, err := app.ParseManifest(manifestPath)
	if err != nil {
		return runCmd
	}

	runCmd.RunE = func(cmd *cobra.Command, args []string) error {
		extensionRoot := args[0]

		extension.Root = extensionRoot
		if err != nil {
			return fmt.Errorf("failed to parse manifest: %s", err)
		}

		rootList := tui.NewRootList(nil, extension)
		model := tui.NewModel(rootList)

		return tui.Draw(model)
	}

	for _, command := range extension.Commands {
		runCmd.AddCommand(NewExtensionSubCommand(extension, command))
	}

	return runCmd
}
