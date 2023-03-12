package cmd

import (
	"fmt"
	"os"
	"path"
	"strconv"

	_ "embed"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/tui"
)

//go:embed sunbeam.json
var defaultManifest string

func exitWithErrorMsg(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg, args...)
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
			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")

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

			runner := tui.NewRunner(generator, configDir)
			model := tui.NewModel(runner, tui.SunbeamOptions{
				Padding:   padding,
				MaxHeight: maxHeight,
			})
			model.Draw()
		},
	}

	rootCmd.PersistentFlags().IntP("padding", "p", lookupInt("SUNBEAM_PADDING", 0), "padding around the window")
	rootCmd.PersistentFlags().IntP("height", "H", lookupInt("SUNBEAM_HEIGHT", 0), "maximum height of the window")
	rootCmd.Flags().StringP("directory", "C", "", "cd to dir before starting sunbeam")

	rootCmd.AddGroup(coreCommandsGroup, extensionCommandsGroup)

	rootCmd.AddCommand(NewRunCmd())
	rootCmd.AddCommand(NewPushCmd())
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewCopyCmd())
	rootCmd.AddCommand(NewOpenCmd())
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

			extraArgs := []string{}
			if len(args) > 1 {
				extraArgs = args[1:]
			}

			cwd, _ := os.Getwd()
			generator := tui.NewCommandGenerator(binPath, extraArgs, cwd)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			runner := tui.NewRunner(generator, cwd)

			tui.NewModel(runner, tui.SunbeamOptions{
				Padding:   0,
				MaxHeight: 0,
			}).Draw()
		},
	}
}

func lookupInt(key string, fallback int) int {
	if env, ok := os.LookupEnv(key); ok {
		if value, err := strconv.Atoi(env); err == nil {
			return value
		}
	}

	return fallback
}
