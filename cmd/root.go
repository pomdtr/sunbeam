package cmd

import (
	"fmt"
	"os"

	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	Version = "dev"
	Date    = "unknown"
)

const (
	CommandGroupCore      = "core"
	CommandGroupDev       = "dev"
	CommandGroupExtension = "extension"
)

func IsSunbeamRunning() bool {
	return len(os.Getenv("SUNBEAM")) > 0
}

func NewRootCmd() (*cobra.Command, error) {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		Version:      fmt.Sprintf("%s (%s)", Version, Date),
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    CommandGroupCore,
		Title: "Core Commands:",
	}, &cobra.Group{
		ID:    CommandGroupDev,
		Title: "Development Commands:",
	})
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewCmdFetch())
	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(NewCmdExtension())
	rootCmd.AddCommand(NewCmdEdit())
	rootCmd.AddCommand(NewCmdCopy())
	rootCmd.AddCommand(NewCmdPaste())
	rootCmd.AddCommand(NewCmdOpen())
	rootCmd.AddCommand(NewCmdConfig())

	docCmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation for sunbeam",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := buildDoc(rootCmd)
			if err != nil {
				return err
			}

			fmt.Println(doc)
			return nil
		},
	}
	rootCmd.AddCommand(docCmd)

	manCmd := &cobra.Command{
		Use:    "generate-man-pages [path]",
		Short:  "Generate Man Pages for sunbeam",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			header := &doc.GenManHeader{
				Title:   "MINE",
				Section: "3",
			}
			err := doc.GenManTree(rootCmd, header, args[0])
			if err != nil {
				return err
			}

			return nil
		},
	}
	rootCmd.AddCommand(manCmd)

	if IsSunbeamRunning() {
		return rootCmd, nil
	}

	extensionMap, err := FindExtensions()
	if err != nil {
		return nil, err
	}

	if len(extensionMap) == 0 {
		return rootCmd, nil
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		config, err := config.Load()
		if err != nil {
			return err
		}

		var rootItems []types.RootItem
		rootItems = append(rootItems, config.Root...)
		for _, extension := range extensionMap {
			rootItems = append(rootItems, extension.RootItems()...)
		}

		rootList := tui.NewRootList("Sunbeam", extensionMap, rootItems...)
		return tui.Draw(rootList)
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    CommandGroupExtension,
		Title: "Extension Commands:",
	})

	for alias, extension := range extensionMap {
		command, err := NewCmdCustom(alias, extension)
		if err != nil {
			return nil, err
		}

		rootCmd.AddCommand(command)
	}

	return rootCmd, nil
}
