package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	_ "embed"

	"github.com/joho/godotenv"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/internal"
)

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

func NewRootCmd() (*cobra.Command, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user home directory: %s", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get current working directory: %s", err)
	}

	configDir := path.Join(homeDir, ".config", "sunbeam")
	dataDir := path.Join(homeDir, ".local", "share", "sunbeam")
	extensionDir := path.Join(dataDir, "extensions")

	configFile := path.Join(configDir, "sunbeam.yml")

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		Args: cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			dotenv := path.Join(cwd, ".env")
			if _, err := os.Stat(dotenv); !os.IsNotExist(err) {
				err := godotenv.Load(dotenv)
				if err != nil {
					return fmt.Errorf("could not load env file: %s", err)
				}
			}

			os.Setenv("SUNBEAM", "1")
			return nil
		},
		// If the config file does not exist, create it
		RunE: func(cmd *cobra.Command, args []string) error {
			var generator func() ([]byte, error)
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				generator = func() ([]byte, error) {
					return io.ReadAll(os.Stdin)
				}
			} else if _, err := os.Stat(configFile); !os.IsNotExist(err) {
				generator = internal.NewFileGenerator(configFile)
			} else {
				return cmd.Usage()
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator()
				if err != nil {
					return fmt.Errorf("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return nil
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()
			return nil
		},
	}

	rootCmd.AddGroup(coreCommandsGroup, extensionCommandsGroup)

	rootCmd.AddCommand(NewCopyCmd())
	rootCmd.AddCommand(NewPasteCmd())
	rootCmd.AddCommand(NewExtensionCmd(extensionDir))
	rootCmd.AddCommand(NewOpenCmd())
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewReadCmd())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewTriggerCmd())
	rootCmd.AddCommand(NewCmdAsk())
	rootCmd.AddCommand(NewCmdEval())

	coreCommands := map[string]struct{}{}
	for _, coreCommand := range rootCmd.Commands() {
		coreCommand.GroupID = coreCommandsGroup.ID
		coreCommands[coreCommand.Name()] = struct{}{}
	}

	// Add the extension commands
	extensions, err := ListExtensions(extensionDir)
	if err != nil {
		return nil, fmt.Errorf("could not list extensions: %s", err)
	}

	for _, extension := range extensions {
		// Skip if the extension name conflicts with a core command
		if _, ok := coreCommands[extension]; ok {
			continue
		}

		rootCmd.AddCommand(NewExtensionShortcutCmd(extensionDir, extension))
	}

	return rootCmd, nil
}

func NewExtensionShortcutCmd(extensionDir string, extensionName string) *cobra.Command {
	return &cobra.Command{
		Use:                extensionName,
		DisableFlagParsing: true,
		GroupID:            extensionCommandsGroup.ID,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			binPath := path.Join(extensionDir, extensionName, extensionBinaryName)
			if _, err := os.Stat(binPath); os.IsNotExist(err) {
				return fmt.Errorf("extension not found: %s", extensionName)
			}

			command := fmt.Sprintf("%s %s", binPath, strings.Join(args, " "))

			cwd, _ := os.Getwd()
			generator := internal.NewCommandGenerator(command, "", cwd)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator()
				if err != nil {
					return fmt.Errorf("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return nil
			}

			runner := internal.NewRunner(generator)

			internal.NewPaginator(runner).Draw()
			return nil
		},
	}
}
