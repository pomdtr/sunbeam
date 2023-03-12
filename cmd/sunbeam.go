package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"

	_ "embed"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/tui"
)

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
	dataDir := path.Join(homeDir, ".local", "share", "sunbeam")
	extensionDir := path.Join(dataDir, "extensions")

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		Args:    cobra.NoArgs,
		Version: version,
		// If the config file does not exist, create it
		Run: func(cmd *cobra.Command, args []string) {
			padding, _ := cmd.Flags().GetInt("padding")
			maxHeight, _ := cmd.Flags().GetInt("height")

			cwd, _ := os.Getwd()
			manifestPath := path.Join(cwd, "sunbeam.json")

			generator := func(string) ([]byte, error) {
				listItems := make([]schemas.ListItem, 0)
				if _, err := os.Stat(manifestPath); err == nil {
					listItems = append(listItems, schemas.ListItem{
						Title: "Read Current Directory Config",
						Actions: []schemas.Action{
							{
								Type:     schemas.ReadAction,
								RawTitle: "Read",
								Page:     manifestPath,
							},
						},
					})
				}

				listItems = append(listItems, schemas.ListItem{
					Title: "Manage Installed Extensions",
					Actions: []schemas.Action{
						{
							Type:      schemas.RunAction,
							RawTitle:  "Manage",
							Command:   "sunbeam extension manage",
							OnSuccess: schemas.PushOnSuccess,
						},
					},
				})

				listItems = append(listItems, schemas.ListItem{
					Title: "Browse Extensions from Github",
					Actions: []schemas.Action{
						{
							Type:      schemas.RunAction,
							RawTitle:  "Browse",
							Command:   "sunbeam extension browse",
							OnSuccess: schemas.PushOnSuccess,
						},
					},
				})

				// Add the core commands
				listItems = append(listItems, schemas.ListItem{
					Title: "Create Extension",
					Actions: []schemas.Action{
						{
							Type:     schemas.RunAction,
							RawTitle: "Create Extension",
							Command:  "sunbeam extension create {{ extensionName }}",
							Inputs: []schemas.FormInput{
								{
									Type:        schemas.TextField,
									Name:        "extensionName",
									Placeholder: "my-extension",
									Title:       "Extension Name",
								},
							},
						},
					},
				})

				page := schemas.Page{
					Type: schemas.ListPage,
					List: &schemas.List{
						Items: listItems,
					},
				}

				return json.Marshal(page)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			runner := tui.NewRunner(generator, cwd)
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
