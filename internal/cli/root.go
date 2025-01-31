package cli

import (
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/history"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
)

func IsSunbeamRunning() bool {
	return len(os.Getenv("SUNBEAM")) > 0
}

func NewRootCmd() (*cobra.Command, error) {
	var flags struct {
		reload bool
	}
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !flags.reload {
				return cmd.Usage()
			}

			exts, err := extensions.LoadExtensions(utils.ExtensionsDir(), false)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("no extensions found")
			} else if err != nil {
				return fmt.Errorf("failed to load extensions: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Reloaded %d extensions\n", len(exts))
			return nil
		},
	}

	rootCmd.Flags().BoolVar(&flags.reload, "reload", false, "Reload extensions manifest")
	rootCmd.AddCommand(NewValidateCmd())

	rootCmd.AddGroup(&cobra.Group{
		ID:    "extension",
		Title: "Extensions Commands:",
	})

	exts, err := extensions.LoadExtensions(utils.ExtensionsDir(), true)
	if errors.Is(err, os.ErrNotExist) {
		return rootCmd, nil
	} else if err != nil {
		return nil, err
	}

	for _, extension := range exts {
		command, err := NewCmdExtension(extension.Name, extension)
		if err != nil {
			return nil, err
		}
		rootCmd.AddCommand(command)
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			return fmt.Errorf("sunbeam must be run in a terminal")
		}

		history, err := history.Load(history.Path)
		if err != nil {
			return err
		}

		rootList := tui.NewRootList(history, func() ([]sunbeam.ListItem, error) {
			exts, err := extensions.LoadExtensions(utils.ExtensionsDir(), true)
			if err != nil {
				return nil, err
			}

			var items []sunbeam.ListItem
			for _, extension := range exts {
				items = append(items, extension.RootItems()...)
			}

			return items, nil
		})
		return tui.Draw(rootList)

	}

	return rootCmd, nil
}
