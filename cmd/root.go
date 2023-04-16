package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	_ "embed"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/internal"
	"github.com/pomdtr/sunbeam/utils"
)

func Draw(generator internal.PageGenerator) error {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
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
	options := internal.SunbeamOptions{
		MaxHeight: utils.LookupInt("SUNBEAM_HEIGHT", 0),
		Padding:   utils.LookupInt("SUNBEAM_PADDING", 0),
	}
	paginator := internal.NewPaginator(runner, options)

	var p *tea.Program
	if options.MaxHeight == 0 {
		p = tea.NewProgram(paginator, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(paginator)
	}

	_, err := p.Run()
	if err != nil {
		return err
	}
	return nil
}

func NewRootCmd() (*cobra.Command, error) {

	dataDir := path.Join(xdg.DataHome, "sunbeam")
	extensionDir := path.Join(dataDir, "extensions")

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "sunbeam",
		Short: "Command Line Launcher",
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		Args: cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			os.Setenv("SUNBEAM", "true")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				return Draw(internal.NewStaticGenerator(os.Stdin))
			}

			return cmd.Usage()
		},
	}

	rootCmd.AddCommand(NewExtensionCmd(extensionDir))
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewPushCmd())
	rootCmd.AddCommand(NewCmdServe())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewTriggerCmd())
	rootCmd.AddCommand(NewCmdRun(extensionDir))

	return rootCmd, nil
}
