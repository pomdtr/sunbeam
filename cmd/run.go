package cmd

import (
	"os"
	"path/filepath"

	"github.com/pomdtr/sunbeam/app"
	"github.com/pomdtr/sunbeam/tui"
	"github.com/spf13/cobra"
)

func NewCmdRun(config *tui.Config) *cobra.Command {
	runCmd := &cobra.Command{
		Use:     "run <extension-root>",
		Short:   "Run commands from an extension in the current directory",
		GroupID: "core",
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

	for _, command := range extension.Commands {
		runCmd.AddCommand(NewExtensionSubCommand(extension, command))
	}

	return runCmd
}
