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
		Use:   "run <extension> [script] [params]",
		Short: "Run a script",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extensionRoot, err := filepath.Abs(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to get absolute path for extension root:", err)
				os.Exit(1)
			}

			manifestPath := filepath.Join(extensionRoot, "sunbeam.yml")
			if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "Extension does not exist")
			}

			extension, err := app.ParseManifest(".", manifestPath)

			rootList := tui.NewRootList(extension.RootItems...)
			model := tui.NewModel(rootList, extension)

			return tui.Draw(model)
		},
	}

	return runCmd
}
