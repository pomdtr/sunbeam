package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/sunbeamlauncher/sunbeam/app"
	"github.com/sunbeamlauncher/sunbeam/tui"
)

func NewCmdRun() *cobra.Command {
	runCmd := &cobra.Command{
		Use:     "run <extension-root>",
		Short:   "Run a extension from a directory",
		GroupID: "core",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			extensionRoot := args[0]
			if fi, err := os.Stat(extensionRoot); err != nil || !fi.IsDir() {
				fmt.Fprintf(os.Stderr, "Directory %s does not exist\n", extensionRoot)
				os.Exit(1)
			}

			manifestPath := filepath.Join(extensionRoot, "sunbeam.yml")
			if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Directory %s is not a sunbeam extension\n", extensionRoot)
				os.Exit(1)
			}

			extension, err := app.ParseManifest(".", manifestPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to parse manifest:", err)
				os.Exit(1)
			}

			rootList := tui.NewRootList(extension.RootItems...)
			model := tui.NewModel(rootList, extension)

			return tui.Draw(model)
		},
	}

	return runCmd
}
