package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"

	_ "embed"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/pomdtr/sunbeam/tui"
)

func exitWithErrorMsg(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}

func Execute(version string) error {
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

			if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
				cmd.Help()
				os.Exit(0)
			}

			generator := func(string) ([]byte, error) {
				bytes, err := os.ReadFile(manifestPath)
				if err != nil {
					return nil, fmt.Errorf("Could not read manifest: %w", err)
				}

				return bytes, nil
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				output, err := generator("")
				if err != nil {
					exitWithErrorMsg("could not generate page: %s", err)
				}

				fmt.Print(string(output))
				return
			}

			runner := tui.NewCommandRunner(generator, &url.URL{
				Scheme: "file",
				Path:   cwd,
			})
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

	rootCmd.AddCommand(NewRunCmd())
	rootCmd.AddCommand(NewPushCmd())
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewCopyCmd())
	rootCmd.AddCommand(NewOpenCmd())
	rootCmd.AddCommand(NewValidateCmd())

	return rootCmd.Execute()
}

func lookupInt(key string, fallback int) int {
	if env, ok := os.LookupEnv(key); ok {
		if value, err := strconv.Atoi(env); err == nil {
			return value
		}
	}

	return fallback
}
