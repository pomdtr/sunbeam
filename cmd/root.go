package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	_ "embed"

	"github.com/joho/godotenv"
	"github.com/mattn/go-isatty"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/tui"
)

//go:embed templates/root.yml
var defaultManifest []byte

//go:embed templates/.env
var defaultEnv []byte

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

func Execute(version string, schema *jsonschema.Schema) error {
	homeDir, _ := os.UserHomeDir()

	configDir := path.Join(homeDir, ".config", "sunbeam")
	dataDir := path.Join(homeDir, ".local", "share", "sunbeam")
	extensionDir := path.Join(dataDir, "extensions")

	rootFile := path.Join(configDir, "root.yml")
	envFile := path.Join(configDir, ".env")
	validator := func(page []byte) error {
		var v any
		if err := json.Unmarshal(page, &v); err != nil {
			return fmt.Errorf("unable to parse input: %s", err)
		}

		return schema.Validate(v)
	}

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		Args:    cobra.NoArgs,
		Version: version,
		PreRun: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat(rootFile); os.IsNotExist(err) {
				// Create the config file
				_ = os.MkdirAll(configDir, 0755)
				// Write the default manifest
				_ = os.WriteFile(rootFile, defaultManifest, 0644)
			}

			if _, err := os.Stat(envFile); os.IsNotExist(err) {
				// Create the config file
				_ = os.MkdirAll(configDir, 0755)
				// Write the default manifest
				_ = os.WriteFile(envFile, defaultEnv, 0644)
			}

			godotenv.Load(envFile)
			os.Setenv("SUNBEAM", "1")
		},
		// If the config file does not exist, create it
		Run: func(cmd *cobra.Command, args []string) {
			generator := tui.NewFileGenerator(rootFile)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			cwd, _ := os.Getwd()
			runner := tui.NewRunner(generator, validator, cwd)
			tui.NewPaginator(runner).Draw()
		},
	}

	rootCmd.AddGroup(coreCommandsGroup, extensionCommandsGroup)

	rootCmd.AddCommand(NewRunCmd(validator))
	rootCmd.AddCommand(NewReadCmd(validator))
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewCopyCmd())
	rootCmd.AddCommand(NewOpenCmd())
	rootCmd.AddCommand(NewValidateCmd(validator))
	rootCmd.AddCommand(NewExtensionCmd(extensionDir, validator))

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

		rootCmd.AddCommand(NewExtensionShortcutCmd(extensionDir, validator, extension))
	}

	return rootCmd.Execute()
}

func NewExtensionShortcutCmd(extensionDir string, validator tui.PageValidator, extensionName string) *cobra.Command {
	return &cobra.Command{
		Use:                extensionName,
		DisableFlagParsing: true,
		GroupID:            extensionCommandsGroup.ID,
		Run: func(cmd *cobra.Command, args []string) {
			binPath := path.Join(extensionDir, extensionName, extensionBinaryName)
			if _, err := os.Stat(binPath); os.IsNotExist(err) {
				exitWithErrorMsg("Extension not found: %s", extensionName)
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

			runner := tui.NewRunner(generator, validator, cwd)

			tui.NewPaginator(runner).Draw()
		},
	}
}
