package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	_ "embed"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/types"
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

func isOutputInteractive() bool {
	return isatty.IsTerminal(os.Stderr.Fd())
}

func Draw(generator internal.PageGenerator) error {
	if !isOutputInteractive() {
		output, err := generator()
		if err != nil {
			return fmt.Errorf("could not generate page: %s", err)
		}

		if err := json.NewEncoder(os.Stdout).Encode(output); err != nil {
			return fmt.Errorf("could not encode page: %s", err)
		}
		return nil
	}

	runner := internal.NewRunner(generator)
	paginator := internal.NewPaginator(runner)
	p := tea.NewProgram(paginator, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
	m, err := p.Run()
	if err != nil {
		return err
	}

	model, ok := m.(*internal.Paginator)
	if !ok {
		return fmt.Errorf("could not convert model to paginator")
	}

	if o := model.Output; o != nil {
		switch o.(type) {
		case string:
			fmt.Print(o)
			return nil
		default:
			if err := json.NewEncoder(os.Stdout).Encode(o); err != nil {
				return err
			}
		}
	}

	return nil
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

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		Args: cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			lipgloss.SetColorProfile(termenv.ANSI)

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
			var input []byte
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				input, _ = io.ReadAll(os.Stdin)
			}

			if len(input) == 0 {
				return cmd.Usage()
			}

			return Draw(func() ([]byte, error) {
				var page types.Page
				if err := json.Unmarshal(input, &page); err != nil {
					return nil, fmt.Errorf("could not unmarshal page: %s", err)
				}

				return json.Marshal(page)
			})

		},
	}

	rootCmd.AddGroup(coreCommandsGroup, extensionCommandsGroup)

	rootCmd.AddCommand(NewCopyCmd())
	rootCmd.AddCommand(NewPasteCmd())
	rootCmd.AddCommand(NewExtensionCmd(extensionDir))
	rootCmd.AddCommand(NewOpenCmd())
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewReadCmd())
	rootCmd.AddCommand(NewCmdServe())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewTriggerCmd())
	rootCmd.AddCommand(NewCmdAsk())
	rootCmd.AddCommand(NewCmdEval())

	return rootCmd, nil
}
