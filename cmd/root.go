package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	_ "embed"

	"github.com/joho/godotenv"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
)

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

type SunbeamConfig struct {
	RootActions []types.Action `yaml:"rootItems"`
}

func NewRootCmd() (*cobra.Command, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user home directory: %s", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get current working directory: %s", err)
	}

	dataDir := path.Join(homeDir, ".local", "share", "sunbeam")
	extensionDir := path.Join(dataDir, "extensions")

	configDir := path.Join(homeDir, ".config", "sunbeam")
	envFile := path.Join(configDir, "sunbeam.env")
	configFile := path.Join(configDir, "sunbeam.yml")

	var config SunbeamConfig
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		bytes, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("could not read config file: %s", err)
		}
		if yaml.Unmarshal(bytes, &config); err != nil {
			return nil, fmt.Errorf("could not parse config file: %s", err)
		}
	}
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		Args: cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat(envFile); !os.IsNotExist(err) {
				err := godotenv.Load(envFile)
				if err != nil {
					exitWithErrorMsg("could not load env file: %s", err)
				}
			}

			dotenv := path.Join(cwd, ".env")
			if _, err := os.Stat(dotenv); !os.IsNotExist(err) {
				err := godotenv.Load(dotenv)
				if err != nil {
					exitWithErrorMsg("could not load env file: %s", err)
				}
			}

			os.Setenv("SUNBEAM", "1")
		},
		// If the config file does not exist, create it
		Run: func(cmd *cobra.Command, args []string) {
			generator := func(string) ([]byte, error) {
				items := make([]types.ListItem, 0)

				for _, ext := range []string{"yaml", "yml", "json"} {
					manifestPath := path.Join(cwd, fmt.Sprintf("sunbeam.%s", ext))
					if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
						items = append(items, types.ListItem{
							Title:    "Read Current Dir Manifest",
							Subtitle: "Sunbeam",
							Actions: []types.Action{
								{Type: types.ReadAction, Title: "Read", Path: manifestPath},
							},
						})

						break
					}
				}

				binaryPath := path.Join(cwd, extensionBinaryName)
				if _, err := os.Stat(binaryPath); !os.IsNotExist(err) {
					items = append(items, types.ListItem{
						Title:    "Run Current Directory Extension",
						Subtitle: "Sunbeam",
						Actions: []types.Action{
							{Type: types.RunAction, Title: "Run", OnSuccess: types.PushOnSuccess, Command: binaryPath},
						},
					})
				}

				items = append(items, []types.ListItem{
					{
						Title:    "Browse Available Extensions",
						Subtitle: "Sunbeam",
						Actions: []types.Action{
							{Type: types.RunAction, Title: "Run", OnSuccess: types.PushOnSuccess, Command: "sunbeam extension browse"},
						},
					},
					{
						Title:    "Manage Extensions",
						Subtitle: "Sunbeam",
						Actions: []types.Action{
							{Type: types.RunAction, Title: "Run", OnSuccess: types.PushOnSuccess, Command: "sunbeam extension manage"},
						},
					},
					{
						Title:    "Create Extension",
						Subtitle: "Sunbeam",
						Actions: []types.Action{
							{
								Type:      types.RunAction,
								Title:     "Run",
								OnSuccess: types.PushOnSuccess,
								Command:   "sunbeam extension create ${input:name}",
								Inputs: []types.Input{
									{Name: "name", Type: types.TextFieldInput, Title: "Extension name", Placeholder: "Extension Name"},
								},
							},
						},
					},
				}...)

				for _, item := range config.RootActions {
					items = append(items, types.ListItem{
						Title:    item.Title,
						Subtitle: "Root Action",
						Actions: []types.Action{
							item,
						},
					})
				}

				return json.Marshal(types.Page{
					Type:  types.ListPage,
					Items: items,
				})
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			runner := internal.NewRunner(generator)
			internal.NewPaginator(runner).Draw()
		},
	}

	rootCmd.AddGroup(coreCommandsGroup, extensionCommandsGroup)

	rootCmd.AddCommand(NewCopyCmd())
	rootCmd.AddCommand(NewExtensionCmd(extensionDir))
	rootCmd.AddCommand(NewFetchCmd())
	rootCmd.AddCommand(NewOpenCmd())
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewReadCmd())
	rootCmd.AddCommand(NewRunCmd())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewTriggerCmd())

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

	return rootCmd, nil
}

func NewExtensionShortcutCmd(extensionDir string, extensionName string) *cobra.Command {
	return &cobra.Command{
		Use:                extensionName,
		DisableFlagParsing: true,
		GroupID:            extensionCommandsGroup.ID,
		Args:               cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			binPath := path.Join(extensionDir, extensionName, extensionBinaryName)
			if _, err := os.Stat(binPath); os.IsNotExist(err) {
				exitWithErrorMsg("Extension not found: %s", extensionName)
			}

			command := fmt.Sprintf("%s %s", binPath, strings.Join(args, " "))

			cwd, _ := os.Getwd()
			generator := internal.NewCommandGenerator(command, "", cwd)

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			runner := internal.NewRunner(generator)

			internal.NewPaginator(runner).Draw()
		},
	}
}
