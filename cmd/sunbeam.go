package cmd

import (
	"fmt"
	"os"
	"path"

	_ "embed"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/tui"
)

//go:embed sunbeam.json
var defaultManifest string

func exitWithErrorMsg(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg, args...)
	fmt.Println()
	os.Exit(1)
}

var (
	coreCommandsGroup = &cobra.Group{
		ID:    "core",
		Title: "Core Commands",
	}
	extensionCommandsGroup = &cobra.Group{
		ID:    "extension",
		Title: "Extension Commands",
	}
)

func Execute(version string) error {
	homeDir, _ := os.UserHomeDir()

	configDir := path.Join(homeDir, ".config", "sunbeam")
	dataDir := path.Join(homeDir, ".local", "share", "sunbeam")
	extensionDir := path.Join(dataDir, "extensions")

	configFile := path.Join(configDir, "sunbeam.json")

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		Args:    cobra.NoArgs,
		Version: version,
		PreRun: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat(configFile); !os.IsNotExist(err) {
				return
			}

			// Create the config file
			_ = os.MkdirAll(configDir, 0755)

			// Write the default manifest
			_ = os.WriteFile(configFile, []byte(defaultManifest), 0644)
		},
		// If the config file does not exist, create it
		Run: func(cmd *cobra.Command, args []string) {
			generator := func(string) ([]byte, error) {
				return os.ReadFile(configFile)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			cwd, _ := os.Getwd()
			runner := tui.NewRunner(generator, cwd)
			tui.NewModel(runner).Draw()
		},
	}

	rootCmd.AddGroup(coreCommandsGroup, extensionCommandsGroup)

	rootCmd.AddCommand(NewRunCmd())
	rootCmd.AddCommand(NewPushCmd())
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewCopyCmd())
	rootCmd.AddCommand(NewOpenCmd())
	rootCmd.AddCommand(NewHttpCmd())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewExtensionCmd(extensionDir))

	coreCommands := map[string]struct{}{}
	for _, coreCommand := range rootCmd.Commands() {
		coreCommand.GroupID = coreCommandsGroup.ID
		coreCommands[coreCommand.Name()] = struct{}{}
	}

	// Add the extension commands
	extensions, err := ListExtensions(extensionDir)
	if err != nil {
		exitWithErrorMsg("could not list extensions: %s", err)
	}

	for _, extension := range extensions {
		// Skip if the extension name conflicts with a core command
		if _, ok := coreCommands[extension]; ok {
			continue
		}

		rootCmd.AddCommand(NewExtensionShortcutCmd(extensionDir, extension))
	}

	return rootCmd.Execute()
}

func NewExtensionShortcutCmd(extensionDir string, extensionName string) *cobra.Command {
	return &cobra.Command{
		Use:                extensionName,
		DisableFlagParsing: true,
		GroupID:            extensionCommandsGroup.ID,
		Run: func(cmd *cobra.Command, args []string) {
			extensionId, err := ExtensionID(extensionName)
			if err != nil {
				exitWithErrorMsg("Invalid extension: %s", err)
			}

			binPath := path.Join(extensionDir, extensionId, extensionId)

			if _, err := os.Stat(binPath); os.IsNotExist(err) {
				exitWithErrorMsg("Extension not found: %s", extensionId)
			}

			cwd, _ := os.Getwd()
			generator := tui.NewCommandGenerator(binPath, args, cwd)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			runner := tui.NewRunner(generator, cwd)

			tui.NewModel(runner).Draw()
		},
	}
}
